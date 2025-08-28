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

var overlayTests = map[string]struct {
	initFiles     map[string]string
	request       func(serverURL string) (*http.Request, error)
	response      string
	status        int
	resultFiles   []string
	validateFiles map[string]string // file path -> expected content
}{
	"get all overlays": {
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays", nil)
		},
		response: `{"testoverlay":{"files":["/email.ww"], "site":false}}`,
	},

	"get one specific overlay": {
		initFiles: map[string]string{
			"/usr/share/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/overlays/testoverlay", nil)
		},
		response: `{"files":["/email.ww"], "site":false}`,
	},

	"get overlay file": {
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

	"update overlay file": {
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

	"create an overlay": {
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPut, serverURL+"/api/overlays/test", nil)
		},
		response: `{"files":null, "site":true}`,
		resultFiles: []string{
			"/var/lib/warewulf/overlays/test",
		},
	},

	"get all overlays after creation": {
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

	"delete overlay file": {
		initFiles: map[string]string{
			"/var/lib/warewulf/overlays/testoverlay/email.ww": sampleTemplate,
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/overlays/testoverlay/file?path=email.ww&force=true", nil)
		},
		response: `{"files":null, "site":true}`,
	},

	"force delete site overlay": {
		initFiles: map[string]string{
			"/var/lib/warewulf/overlays/test/": "",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/overlays/test?force=true", nil)
		},
		response: `{"files":[], "site":true}`,
	},

	"force delete distribution overlay": {
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
	for name, tt := range overlayTests {
		t.Run(name, func(t *testing.T) {
			warewulfd.SetNoDaemon()
			env := testenv.New(t)
			defer env.RemoveAll()

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
				ja.Assertf(string(body), tt.response) //nolint:govet // tt.response is used as a format string with special tokens
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
