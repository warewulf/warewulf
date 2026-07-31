package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var ipxeHandlerTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{
		"ipxe with NetDevs, KernelVersion, and Authority",
		"/ipxe/00:00:00:00:00:ff",
		"1.1.1 ifname=net:00:00:00:00:00:ff  10.10.0.1 fd00:10::1 10.10.0.1:9873",
		200,
		"10.10.10.12:9873",
	},
	{
		"ipxe over ipv6",
		"/ipxe/00:00:00:00:00:ff",
		"1.1.1 ifname=net:00:00:00:00:00:ff  10.10.0.1 fd00:10::1 [fd00:10::1]:9873",
		200,
		"[fd00:10::10:12]:9873",
	},
}

var ipxeDracutNetTests = []struct {
	description string
	url         string
	expected    string
	unexpected  string
}{
	{
		"IPv6-only node gets IPv6 autoconfiguration",
		"/ipxe/00:00:00:00:00:06",
		"ip=net:auto6",
		"ip=net:dhcp",
	},
	{
		"IPv4-only node keeps DHCPv4",
		"/ipxe/00:00:00:00:00:04",
		"ip=net:dhcp",
		"ip=net:auto6",
	},
	{
		"dual-stack node keeps DHCPv4",
		"/ipxe/00:00:00:00:00:46",
		"ip=net:dhcp",
		"ip=net:auto6",
	},
	{
		"node with no address configured keeps DHCPv4",
		"/ipxe/00:00:00:00:00:00",
		"ip=net:dhcp",
		"ip=net:auto6",
	},
}

// dracutNetLine returns the "set dracut_net" line from the given iPXE entry.
// default.ipxe defines dracut_net twice, once under :dracut_continue and once
// under :dracut_static_continue, and only the former is under test: the latter
// builds its ip= argument from the IPv4 fields. Anchoring on the label keeps the
// assertions on the intended entry even if the entries are reordered.
func dracutNetLine(script string, label string) string {
	inEntry := false
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(line, ":") {
			inEntry = line == label
			continue
		}
		if inEntry && strings.HasPrefix(line, "set dracut_net ") {
			return line
		}
	}
	return ""
}

// The :dracut entry must not request DHCPv4 for a device that only has an IPv6
// address: NetworkManager's initrd generator turns ip=<device>:dhcp into
// ipv4.method=auto with ipv4.may-fail=no, which can never activate on a fabric
// with no DHCPv4 server, so the initramfs never fetches the image.
func Test_HandleIpxeDracutNet(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("/etc/warewulf/nodes.conf", `nodes:
  n6:
    image name: rockylinux-9
    network devices:
      default:
        hwaddr: 00:00:00:00:00:06
        device: net
        ipaddr6: fd00:10::6
        prefixlen6: "64"
  n4:
    image name: rockylinux-9
    network devices:
      default:
        hwaddr: 00:00:00:00:00:04
        device: net
        ipaddr: 10.10.0.4
        netmask: 255.255.255.0
  n46:
    image name: rockylinux-9
    network devices:
      default:
        hwaddr: 00:00:00:00:00:46
        device: net
        ipaddr: 10.10.0.46
        netmask: 255.255.255.0
        ipaddr6: fd00:10::46
        prefixlen6: "64"
  n0:
    image name: rockylinux-9
    network devices:
      default:
        hwaddr: 00:00:00:00:00:00
        device: net`)
	env.ImportFile("etc/warewulf/ipxe/default.ipxe", "../../../etc/ipxe/default.ipxe")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	// an IPv6-only server, as deployed on the fabric where this was found
	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse
	conf.Ipaddr = ""
	conf.Ipaddr6 = "fd00:10::1"

	for _, tt := range ipxeDracutNetTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = "[fd00:10::10:12]:9873"
			w := httptest.NewRecorder()
			HandleIpxe(w, req)
			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			data, readErr := io.ReadAll(res.Body)
			assert.NoError(t, readErr)
			assert.Equal(t, 200, res.StatusCode)

			dracutNet := dracutNetLine(string(data), ":dracut_continue")
			assert.NotEmpty(t, dracutNet, "template did not render a dracut_net line")
			assert.Contains(t, dracutNet, tt.expected)
			assert.NotContains(t, dracutNet, tt.unexpected)
		})
	}
}

func Test_HandleIpxe(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("/etc/warewulf/nodes.conf", `nodes:
  n3:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:ff
        device: net
    ipxe template: test
    kernel:
      version: 1.1.1`)

	env.WriteFile("/etc/warewulf/ipxe/test.ipxe", "{{.KernelVersion}}{{range $devname, $netdev := .NetDevs}}{{if and $netdev.Hwaddr $netdev.Device}} ifname={{$netdev.Device}}:{{$netdev.Hwaddr}} {{end}}{{end}} {{.Ipaddr}} {{.Ipaddr6}} {{.Authority}}")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse
	conf.Ipaddr = "10.10.0.1"
	conf.Ipaddr6 = "fd00:10::1"

	for _, tt := range ipxeHandlerTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleIpxe(w, req)
			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			data, readErr := io.ReadAll(res.Body)
			assert.NoError(t, readErr)
			if tt.body != "" {
				assert.Equal(t, tt.body, string(data))
			}
			assert.Equal(t, tt.status, res.StatusCode)
		})
	}
}
