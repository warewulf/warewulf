package api

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

var profileTests = map[string]struct {
	initConf   string
	initFiles  []string
	request    func(serverURL string) (*http.Request, error)
	response   string
	status     int
	resultConf string
}{
	"get all profiles": {
		initConf: `
nodeprofiles:
  default: {}
nodes: {}
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/profiles", nil)
		},
		response: `{"default": {}}`,
	},

	"add a new profile": {
		initConf: `
nodeprofiles:
  default: {}
nodes: {}
`,
		request: func(serverURL string) (*http.Request, error) {
			testProfile := `{"profile": {"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`
			return http.NewRequest(http.MethodPut, serverURL+"/api/profiles/p1", bytes.NewBuffer([]byte(testProfile)))
		},
		response: `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
	},

	"test idempotency (replacing existing profile)": {
		initConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
		request: func(serverURL string) (*http.Request, error) {
			testProfile := `{"profile": {"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`
			return http.NewRequest(http.MethodPut, serverURL+"/api/profiles/p1", bytes.NewBuffer([]byte(testProfile)))
		},
		response: `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
	},

	"test preventing replacing a profile": {
		initConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
		request: func(serverURL string) (*http.Request, error) {
			testProfile := `{"profile": {"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`
			req, err := http.NewRequest(http.MethodPut, serverURL+"/api/profiles/p1", bytes.NewBuffer([]byte(testProfile)))
			req.Header.Set("If-None-Match", "*")
			return req, err
		},
		response: `{"error": "invalid argument: profile 'p1' already exists", "status": "INVALID_ARGUMENT"}`,
		status:   http.StatusBadRequest,
		resultConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
	},

	"get one specific profile": {
		initConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/profiles/p1", nil)
		},
		response: `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
	},

	"update a profile": {
		initConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
		request: func(serverURL string) (*http.Request, error) {
			updateProfile := `{"profile": {"kernel": {"version": "v1.0.1-newversion"}}}`
			return http.NewRequest(http.MethodPatch, serverURL+"/api/profiles/p1", bytes.NewBuffer([]byte(updateProfile)))
		},
		response: `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.1-newversion"
      args:
        - "kernel-args"
nodes: {}
`,
	},

	"test delete a profile": {
		initConf: `
nodeprofiles:
  default: {}
  p1:
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
nodes: {}
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/profiles/p1", nil)
		},
		response: `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles:
  default: {}
nodes: {}
`,
	},
}

func TestProfileAPI(t *testing.T) {
	for name, tt := range profileTests {
		t.Run(name, func(t *testing.T) {
			warewulfd.SetNoDaemon()
			env := testenv.New(t)
			defer env.RemoveAll()

			env.WriteFile("/etc/warewulf/nodes.conf", tt.initConf)
			for _, fileName := range tt.initFiles {
				env.CreateFile(fileName)
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

			ja := jsonassert.New(t)
			ja.Assertf(string(body), tt.response) //nolint:govet

			if tt.resultConf != "" {
				assert.YAMLEq(t, tt.resultConf, env.ReadFile("/etc/warewulf/nodes.conf"))
			}
		})
	}
}
