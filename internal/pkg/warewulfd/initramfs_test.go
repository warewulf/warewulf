package warewulfd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var initramfsHandlerTests = []struct {
	description string
	url         string
	status      int
	ip          string
}{
	{"find initramfs", "/initramfs/00:00:00:ff:ff:ff", 200, "10.10.10.10:9873"},
}

func Test_HandleInitramfs(t *testing.T) {
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
    - default`)

	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/boot/vmlinuz-1.1.0")
	env.CreateFile("/var/lib/warewulf/chroots/suse/rootfs/boot/initramfs-1.1.0.img")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse

	for _, tt := range initramfsHandlerTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleInitramfs(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.status, res.StatusCode)
		})
	}
}
