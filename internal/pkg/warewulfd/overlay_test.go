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

var overlaySendTests = map[string]struct {
	url    string
	body   string
	status int
}{
	"get file": {
		url:    "/overlay-file/pub/non-template",
		body:   "Non-template: {{.Id}}",
		status: 200,
	},
	"getting a missing file returns 404": {
		url:    "/overlay-file/pub/does-not-exist",
		body:   "",
		status: 404,
	},
	"get raw template": {
		url:    "/overlay-file/pub/template.ww",
		body:   "Template: {{.Id}}",
		status: 200,
	},
	"get rendered template": {
		url:    "/overlay-file/pub/template.ww?render=n1",
		body:   "Template: n1",
		status: 200,
	},
	"get rendered template without explicit suffix": {
		url:    "/overlay-file/pub/template?render=n1",
		body:   "Template: n1",
		status: 200,
	},
	"explicit suffix required when no node specified": {
		url:    "/overlay-file/pub/template",
		body:   "",
		status: 404,
	},
	"getting a template with a missing node returns 404": {
		url:    "/overlay-file/pub/test.template.ww?render=n2",
		body:   "",
		status: 404,
	},
	"don't render non-template files": {
		url:    "/overlay-file/pub/non-template?render=n1",
		body:   "Non-template: {{.Id}}",
		status: 200,
	},
	"getting a missing template returns 404": {
		url:    "/overlay-file/pub/does-not-exist.ww?render=n1",
		body:   "",
		status: 404,
	},
	"get a file from a subdir": {
		url:    "/overlay-file/pub/subdir/non-template",
		body:   "Non-template (subdir): {{.Id}}",
		status: 200,
	},
	"render a template from a subdir": {
		url:    "/overlay-file/pub/subdir/template.ww?render=n1",
		body:   "Template (subdir): n1",
		status: 200,
	},
}

func Test_OverlaySend(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile("etc/warewulf/warewulf.conf", `
warewulf:
  secure: false
`)
	env.WriteFile("etc/warewulf/nodes.conf", `
nodeprofiles:
  default: {}
nodes:
  n1: {}
`)
	_ = env.Configure()
	env.WriteFile("var/lib/warewulf/overlays/pub/rootfs/non-template", "Non-template: {{.Id}}")
	env.WriteFile("var/lib/warewulf/overlays/pub/rootfs/template.ww", "Template: {{.Id}}")
	env.WriteFile("var/lib/warewulf/overlays/pub/rootfs/subdir/non-template", "Non-template (subdir): {{.Id}}")
	env.WriteFile("var/lib/warewulf/overlays/pub/rootfs/subdir/template.ww", "Template (subdir): {{.Id}}")

	for description, tt := range overlaySendTests {
		t.Run(description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			HandleOverlayFile(w, req)
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

var systemOverlayTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"system overlay", "/system/00:00:00:ff:ff:ff", "system overlay", 200, "10.10.10.10:9873"},
	{"fake overlay returns 404", "/system/00:00:00:ff:ff:ff?overlay=fake", "", 404, "10.10.10.10:9873"},
	{"specific overlay", "/system/00:00:00:ff:ff:ff?overlay=o1", "specific overlay", 200, "10.10.10.10:9873"},
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
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse

	assert.NoError(t, os.MkdirAll(path.Join(conf.Paths.OverlayProvisiondir(), "n1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__SYSTEM__.img"), []byte("system overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__RUNTIME__.img"), []byte("runtime overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "o1.img"), []byte("specific overlay"), 0600))

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
