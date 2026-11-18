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

var systemOverlayTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"system overlay", "/system/00:00:00:ff:ff:ff", "system overlay", 200, "10.10.10.10:9873"},
}

var runtimeOverlayTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"runtime overlay", "/runtime/00:00:00:ff:ff:ff", "runtime overlay", 200, "10.10.10.10:9873"},
}

func Test_HandleSystemRuntimeOverlay(t *testing.T) {
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

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	conf.Warewulf.SecureP = warewulfconf.NewSecureRoutes()

	assert.NoError(t, os.MkdirAll(path.Join(conf.Paths.OverlayProvisiondir(), "n1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__SYSTEM__.img"), []byte("system overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__RUNTIME__.img"), []byte("runtime overlay"), 0600))

	for _, tt := range systemOverlayTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleSystemOverlay(w, req)
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

	for _, tt := range runtimeOverlayTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleRuntimeOverlay(w, req)
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

func Test_HandleOverlay_SecureRoutes(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("/etc/warewulf/nodes.conf", `nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:ff:ff:ff`)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	assert.NoError(t, os.MkdirAll(path.Join(conf.Paths.OverlayProvisiondir(), "n1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__SYSTEM__.img"), []byte("system overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__RUNTIME__.img"), []byte("runtime overlay"), 0600))

	tests := map[string]struct {
		secure        *warewulfconf.SecureRoutes
		systemStatus  int
		runtimeStatus int
	}{
		"nil":     {nil, http.StatusOK, http.StatusForbidden},
		"none":    {warewulfconf.NewSecureRoutes(), http.StatusOK, http.StatusOK},
		"system":  {warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteSystem), http.StatusForbidden, http.StatusOK},
		"runtime": {warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteRuntime), http.StatusOK, http.StatusForbidden},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conf.Warewulf.SecureP = tt.secure
			for _, port := range []string{"1234", "987"} {
				systemStatus, runtimeStatus := tt.systemStatus, tt.runtimeStatus
				if port == "987" {
					systemStatus, runtimeStatus = http.StatusOK, http.StatusOK
				}

				req := httptest.NewRequest(http.MethodGet, "/system/00:00:00:ff:ff:ff", nil)
				req.RemoteAddr = "10.10.10.10:" + port
				w := httptest.NewRecorder()
				HandleSystemOverlay(w, req)
				assert.Equal(t, systemStatus, w.Result().StatusCode, "system overlay from port %s", port)

				req = httptest.NewRequest(http.MethodGet, "/runtime/00:00:00:ff:ff:ff", nil)
				req.RemoteAddr = "10.10.10.10:" + port
				w = httptest.NewRecorder()
				HandleRuntimeOverlay(w, req)
				assert.Equal(t, runtimeStatus, w.Result().StatusCode, "runtime overlay from port %s", port)
			}
		})
	}
}
