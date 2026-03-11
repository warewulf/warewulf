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

var grubHandlerTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"grub config for node with image", "/grub/00:00:00:ff:ff:ff", "", 200, "10.10.10.10:9873"},
	{"grub config rendered with tag", "/grub/00:00:00:00:ff:ff", "dracut 10.10.0.1:9873", 200, "10.10.10.11:9873"},
}

func Test_HandleGrub(t *testing.T) {
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

	env.WriteFile("/etc/warewulf/grub/grub.cfg.ww", "{{ .Tags.GrubMenuEntry }} {{ .Authority }}")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse
	conf.Ipaddr = "10.10.0.1"
	conf.Ipaddr6 = "fd00:10::1"

	for _, tt := range grubHandlerTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleGrub(w, req)
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
