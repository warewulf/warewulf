package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var efiBootTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"find shim", "/efiboot/shim.efi", "", 200, "10.10.10.10:9873"},
	{"find shim: node with missing image returns 404", "/efiboot/shim.efi", "", 404, "10.10.10.11:9873"},
	{"find grub", "/efiboot/grub.efi", "", 200, "10.10.10.10:9873"},
	{"find grub: node with missing image returns 404", "/efiboot/grub.efi", "", 404, "10.10.10.11:9873"},
	{"find grub.cfg", "/efiboot/grub.cfg", "dracut 10.10.0.1:9873", 200, "10.10.10.11:9873"},
}

var efiIpxeTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	// iPXE binaries are served from the iPXE source directory.
	{"find ipxe binary", "/efiboot/ipxe-snponly-x86_64.efi", "IPXE-SNPONLY-X86", 200, "10.10.10.10:9873"},
	{"find nested arm64 ipxe binary", "/efiboot/snponly.efi", "IPXE-SNPONLY-ARM64", 200, "10.10.10.10:9873"},
	// iPXE binaries are served without node identification.
	{"find ipxe binary: client without ARP entry or wwid", "/efiboot/ipxe-snponly-x86_64.efi", "IPXE-SNPONLY-X86", 200, "10.10.10.99:9873"},
	{"find ipxe binary: IPv6 client without wwid", "/efiboot/snponly.efi", "IPXE-SNPONLY-ARM64", 200, "[fd00::5]:1234"},
	// Absolute-path IpxeBinaries values are not servable over HTTP.
	{"absolute path ipxe binary not served", "/efiboot/ipxe-abs.efi", "", 404, "10.10.10.10:9873"},
	{"unknown efiboot file", "/efiboot/unknown.efi", "", 404, "10.10.10.10:9873"},
	// Other efiboot files still require node identification.
	{"shim.efi: unidentified client", "/efiboot/shim.efi", "", 400, "10.10.10.99:9873"},
}

func Test_HandleEfiBootIpxe(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("/etc/warewulf/nodes.conf", `nodeprofiles:
  default:
    image name: suse
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:ff:ff:ff
    profiles:
    - default
  d1:
    discoverable: true
    profiles:
    - default`)

	env.WriteFile("/var/tmp/arpcache", `IP address       HW type     Flags       HW address            Mask     Device
10.10.10.10    0x1         0x2         00:00:00:ff:ff:ff     *        dummy`)
	prevArpFile := arpFile
	arpFile = env.GetPath("/var/tmp/arpcache")
	defer func() {
		arpFile = prevArpFile
	}()

	env.MkdirAll("/var/lib/warewulf/ipxe/arm64-efi")
	env.WriteFile("/var/lib/warewulf/ipxe/ipxe-snponly-x86_64.efi", "IPXE-SNPONLY-X86")
	env.WriteFile("/var/lib/warewulf/ipxe/arm64-efi/snponly.efi", "IPXE-SNPONLY-ARM64")
	// The absolute-path IpxeBinaries value points at this file; serving it
	// would resolve to <Ipxesource>/etc/warewulf/ipxe-abs.efi.
	env.MkdirAll("/var/lib/warewulf/ipxe/etc/warewulf")
	env.WriteFile("/var/lib/warewulf/ipxe/etc/warewulf/ipxe-abs.efi", "IPXE-ABS")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse
	prevIpxesource := conf.Paths.Ipxesource
	prevIpxeBinaries := conf.TFTP.IpxeBinaries
	conf.Paths.Ipxesource = env.GetPath("/var/lib/warewulf/ipxe")
	conf.TFTP.IpxeBinaries = map[string]string{
		"00:07": "ipxe-snponly-x86_64.efi",
		"00:0B": "arm64-efi/snponly.efi",
		"00:99": "/etc/warewulf/ipxe-abs.efi",
	}
	defer func() {
		conf.Paths.Ipxesource = prevIpxesource
		conf.TFTP.IpxeBinaries = prevIpxeBinaries
	}()

	for _, tt := range efiIpxeTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleEfiBoot(w, req)
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

	// An anonymous iPXE fetch (no ARP entry, no wwid) must not trigger node
	// discovery: the discoverable node stays undiscovered.
	nodesConfBefore, readErr := os.ReadFile(env.GetPath("/etc/warewulf/nodes.conf"))
	assert.NoError(t, readErr)

	req := httptest.NewRequest(http.MethodGet, "/efiboot/ipxe-snponly-x86_64.efi", nil)
	req.RemoteAddr = "10.10.10.99:9873"
	w := httptest.NewRecorder()
	HandleEfiBoot(w, req)
	res := w.Result()
	defer func() { _ = res.Body.Close() }()
	assert.Equal(t, 200, res.StatusCode)

	nodesConfAfter, readErr := os.ReadFile(env.GetPath("/etc/warewulf/nodes.conf"))
	assert.NoError(t, readErr)
	assert.Equal(t, string(nodesConfBefore), string(nodesConfAfter))
	assert.Contains(t, string(nodesConfAfter), "discoverable: true")
	assert.NoError(t, LoadNodeDB())
}

func Test_HandleEfiBoot(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("/etc/warewulf/nodes.conf", `nodeprofiles:
  default:
    image name: suse
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:ff:ff:ff
    profiles:
    - default
  n2:
    network devices:
      default:
        hwaddr: 00:00:00:00:ff:ff
    image name: none
    tags:
      GrubMenuEntry: dracut`)

	env.WriteFile("/var/tmp/arpcache", `IP address       HW type     Flags       HW address            Mask     Device
10.10.10.10    0x1         0x2         00:00:00:ff:ff:ff     *        dummy
10.10.10.11    0x1         0x2         00:00:00:00:ff:ff     *        dummy`)
	prevArpFile := arpFile
	arpFile = env.GetPath("/var/tmp/arpcache")
	defer func() {
		arpFile = prevArpFile
	}()

	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/usr/lib64/efi/shim.efi")
	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/usr/share/efi/x86_64/grub.efi")
	env.WriteFile("/etc/warewulf/grub/grub.cfg.ww", "{{ .Tags.GrubMenuEntry }} {{ .Authority }}")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse
	conf.Ipaddr = "10.10.0.1"
	conf.Ipaddr6 = "fd00:10::1"

	for _, tt := range efiBootTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleEfiBoot(w, req)
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
