package api

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

const sampleTemplate = `{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}
`

const multiFileTemplate = `{{- range $name := list "alpha" "beta" }}
{{- file $name -}}
content of {{ $name }}
{{ end -}}`

const abortTemplate = `{{- abort -}}`

const abortWithContentTemplate = `some content
{{- abort -}}`

const symlinkTemplate = `{{- softlink "/usr/share/zoneinfo/UTC" -}}`

const multiSymlinkTemplate = `{{- file "link1" -}}
{{- softlink "/target1" -}}
{{- file "link2" -}}
{{- softlink "/target2" -}}`

const sampleNodesConf = `
nodeprofiles: {}
nodes:
  node1: {}
`

var overlayTests = []struct {
	name          string
	initFiles     map[string]string
	request       func(serverURL string) (*http.Request, error)
	response      string
	status        int
	resultFiles   []string
	validateFiles map[string]string // file path -> expected content
}{
	{
		name: "get all overlays",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays", nil)
		},
		response: `{"testoverlay":{"files":["/email.ww"], "site":false}}`,
	},
	{
		name: "get one specific overlay",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/testoverlay", nil)
		},
		response: `{"files":["/email.ww"], "site":false}`,
	},
	{
		name: "get overlay file",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/testoverlay/file?path=email.ww", nil)
		},
		response: `{
			"overlay": "testoverlay",
			"path": "email.ww",
			"contents": "{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}\n",
			"perms": "<<PRESENCE>>",
			"uid": "<<PRESENCE>>",
			"gid": "<<PRESENCE>>"
		}`,
	},
	{
		name: "update overlay file",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPut, serverURL+"/api/overlays/testoverlay/file?path=email.ww", bytes.NewReader([]byte("{\"content\":\"hello world\"}")))
		},
		response: `{"files":["/email.ww"], "site":true}`,
		validateFiles: map[string]string{
			"/var/lib/warewulf/overlays/testoverlay/email.ww": "hello world",
		},
	},
	{
		name: "create an overlay",
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPut, serverURL+"/api/overlays/test", nil)
		},
		response: `{"files":null, "site":true}`,
		resultFiles: []string{
			"/var/lib/warewulf/overlays/test",
		},
	},
	{
		name: "create an overlay conflict",
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPut, serverURL+"/api/overlays/test", nil)
		},
		status: 409,
	},
	{
		name: "get all overlays after creation",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
			"/var/lib/warewulf/overlays/test/rootfs/":           "",
			"/var/lib/warewulf/overlays/testoverlay/email.ww":   sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays", nil)
		},
		response: `{"test":{"files":null, "site":true},"testoverlay":{"files":["/email.ww"], "site":true}}`,
	},
	{
		name: "render multi-file overlay template",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/multioverlay/multi.ww": multiFileTemplate,
			"/etc/warewulf/nodes.conf":                           sampleNodesConf,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/multioverlay/file?path=multi.ww&render=node1", nil)
		},
		response: `{
			"overlay": "multioverlay",
			"path": "multi.ww",
			"contents": "Filename: alpha\ncontent of alpha\nFilename: beta\ncontent of beta\n",
			"perms": "<<PRESENCE>>",
			"uid": "<<PRESENCE>>",
			"gid": "<<PRESENCE>>"
		}`,
	},
	{
		name: "render aborted template",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/multioverlay/abort.ww": abortTemplate,
			"/etc/warewulf/nodes.conf":                           sampleNodesConf,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/multioverlay/file?path=abort.ww&render=node1", nil)
		},
		response: `{
			"overlay": "multioverlay",
			"path": "abort.ww",
			"contents": "Aborted\n",
			"perms": "<<PRESENCE>>",
			"uid": "<<PRESENCE>>",
			"gid": "<<PRESENCE>>"
		}`,
	},
	{
		name: "render aborted template with pre-abort content",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/multioverlay/abort-content.ww": abortWithContentTemplate,
			"/etc/warewulf/nodes.conf":                                   sampleNodesConf,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/multioverlay/file?path=abort-content.ww&render=node1", nil)
		},
		response: `{
			"overlay": "multioverlay",
			"path": "abort-content.ww",
			"contents": "some contentAborted\n",
			"perms": "<<PRESENCE>>",
			"uid": "<<PRESENCE>>",
			"gid": "<<PRESENCE>>"
		}`,
	},
	{
		name: "render single-file symlink template",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/multioverlay/symlink.ww": symlinkTemplate,
			"/etc/warewulf/nodes.conf":                             sampleNodesConf,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/multioverlay/file?path=symlink.ww&render=node1", nil)
		},
		response: `{
			"overlay": "multioverlay",
			"path": "symlink.ww",
			"contents": "Symlink: /usr/share/zoneinfo/UTC\n",
			"perms": "<<PRESENCE>>",
			"uid": "<<PRESENCE>>",
			"gid": "<<PRESENCE>>"
		}`,
	},
	{
		name: "render multi-file symlink template",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/multioverlay/multi-symlink.ww": multiSymlinkTemplate,
			"/etc/warewulf/nodes.conf":                                   sampleNodesConf,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/multioverlay/file?path=multi-symlink.ww&render=node1", nil)
		},
		response: `{
			"overlay": "multioverlay",
			"path": "multi-symlink.ww",
			"contents": "Filename: link1\nSymlink: /target1\nFilename: link2\nSymlink: /target2\n",
			"perms": "<<PRESENCE>>",
			"uid": "<<PRESENCE>>",
			"gid": "<<PRESENCE>>"
		}`,
	},
	{
		name: "delete overlay file",
		initFiles: map[string]string{
			"/var/lib/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/overlays/testoverlay/file?path=email.ww&force=true", nil)
		},
		response: `{"files":null, "site":true}`,
	},
	{
		name: "force delete site overlay",
		initFiles: map[string]string{
			"/var/lib/warewulf/overlays/test/": "",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/overlays/test?force=true", nil)
		},
		response: `{"files":[], "site":true}`,
	},
	{
		name: "force delete distribution overlay",
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/test/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/overlays/test?force=true", nil)
		},
		status: 400,
	},
}

func TestOverlayAPI(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()

	for _, tt := range overlayTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test files
			for fileName, fileContent := range tt.initFiles {
				if strings.HasSuffix(fileName, "/") {
					env.MkdirAll(fileName)
				} else {
					env.WriteFile(fileName, fileContent)
				}
			}

			allowedNets := []net.IPNet{
				{
					IP:   net.IPv4(127, 0, 0, 0),
					Mask: net.CIDRMask(8, 32),
				},
			}
			srv := httptest.NewServer(Handler(nil, allowedNets))
			defer srv.Close()

			req, err := tt.request(srv.URL)
			assert.NoError(t, err)

			resp, err := http.DefaultTransport.RoundTrip(req)
			assert.NoError(t, err)

			expectedStatus := tt.status
			if expectedStatus == 0 {
				expectedStatus = http.StatusOK
			}
			assert.Equal(t, expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.NoError(t, resp.Body.Close())

			if tt.response != "" {
				ja := jsonassert.New(t)
				ja.Assert(string(body), tt.response)
			}

			for _, fileName := range tt.resultFiles {
				assert.DirExists(t, env.GetPath(fileName))
			}

			for filePath, expectedContent := range tt.validateFiles {
				actualContent := env.ReadFile(filePath)
				assert.Equal(t, expectedContent, actualContent)
			}
		})
	}
}
