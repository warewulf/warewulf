package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
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
