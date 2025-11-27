package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var provisionSendTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"system overlay", "/overlay-system/00:00:00:ff:ff:ff", "system overlay", 200, "10.10.10.10:9873"},
	{"runtime overlay", "/overlay-runtime/00:00:00:ff:ff:ff", "runtime overlay", 200, "10.10.10.10:9873"},
	{"fake overlay", "/overlay-system/00:00:00:ff:ff:ff?overlay=fake", "", 404, "10.10.10.10:9873"},
	{"specific overlay", "/overlay-system/00:00:00:ff:ff:ff?overlay=o1", "specific overlay", 200, "10.10.10.10:9873"},
	{"find shim", "/efiboot/shim.efi", "", 200, "10.10.10.10:9873"},
	{"find shim", "/efiboot/shim.efi", "", 404, "10.10.10.11:9873"},
	{"find grub", "/efiboot/grub.efi", "", 200, "10.10.10.10:9873"},
	{"find grub", "/efiboot/grub.efi", "", 404, "10.10.10.11:9873"},
	{"find initramfs", "/provision/00:00:00:ff:ff:ff?stage=initramfs", "", 200, "10.10.10.10:9873"},
	{"ipxe test with NetDevs, KernelVersion, and Authority", "/provision/00:00:00:00:00:ff?stage=ipxe", "1.1.1 ifname=net:00:00:00:00:00:ff  10.10.0.1 fd00:10::1 10.10.0.1:9873", 200, "10.10.10.12:9873"},
	{"ipxe ipv6", "/provision/00:00:00:00:00:ff?stage=ipxe", "1.1.1 ifname=net:00:00:00:00:00:ff  10.10.0.1 fd00:10::1 [fd00:10::1]:9873", 200, "[fd00:10::10:12]:9873"},
	{"find grub.cfg", "/efiboot/grub.cfg", "dracut 10.10.0.1:9873", 200, "10.10.10.11:9873"},
}

func Test_ProvisionSend(t *testing.T) {
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
      GrubMenuEntry: dracut
  n3:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:ff
        device: net
    ipxe template: test
    kernel:
      version: 1.1.1`)

	// create a  arp file as for grub we look up the ip address through the arp cache
	env.WriteFile("/var/tmp/arpcache", `IP address       HW type     Flags       HW address            Mask     Device
10.10.10.10    0x1         0x2         00:00:00:ff:ff:ff     *        dummy
10.10.10.11    0x1         0x2         00:00:00:00:ff:ff     *        dummy
10.10.10.12    0x1         0x2         00:00:00:00:00:ff     *        dummy`)
	prevArpFile := arpFile
	arpFile = env.GetPath("/var/tmp/arpcache")
	defer func() {
		arpFile = prevArpFile
	}()
	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/boot/vmlinuz-1.1.0")
	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/usr/lib64/efi/shim.efi")
	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/usr/share/efi/x86_64/grub.efi")
	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/boot/initramfs-1.1.0.img")
	env.WriteFile("/etc/warewulf/ipxe/test.ipxe", "{{.KernelVersion}}{{range $devname, $netdev := .NetDevs}}{{if and $netdev.Hwaddr $netdev.Device}} ifname={{$netdev.Device}}:{{$netdev.Hwaddr}} {{end}}{{end}} {{.Ipaddr}} {{.Ipaddr6}} {{.Authority}}")
	env.WriteFile("/etc/warewulf/grub/grub.cfg.ww", "{{ .Tags.GrubMenuEntry }} {{ .Authority }}")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse
	conf.Ipaddr = "10.10.0.1"
	conf.Ipaddr6 = "fd00:10::1"
	assert.NoError(t, os.MkdirAll(path.Join(conf.Paths.OverlayProvisiondir(), "n1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__SYSTEM__.img"), []byte("system overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__RUNTIME__.img"), []byte("runtime overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "o1.img"), []byte("specific overlay"), 0600))

	for _, tt := range provisionSendTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			ProvisionSend(w, req)
			res := w.Result()
			defer res.Body.Close()

			data, readErr := io.ReadAll(res.Body)
			assert.NoError(t, readErr)
			if tt.body != "" {
				assert.Equal(t, tt.body, string(data))
			}
			assert.Equal(t, tt.status, res.StatusCode)
		})
	}
}
