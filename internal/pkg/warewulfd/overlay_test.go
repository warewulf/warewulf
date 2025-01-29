package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

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
			OverlaySend(w, req)
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
