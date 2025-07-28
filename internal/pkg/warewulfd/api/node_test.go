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

var nodeTests = map[string]struct {
	initConf    string
	initFiles   []string
	request     func(serverURL string) (*http.Request, error)
	response    string
	status      int
	resultConf  string
	resultFiles []string
}{
	"add a node": {
		initConf: "",
		initFiles: []string{
			"/var/lib/warewulf/overlays/so1/rootfs/file",
			"/var/lib/warewulf/overlays/ro1/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			testNode := `{
  "node":{
    "system overlay": ["so1"],
	"runtime overlay": ["ro1"],
    "kernel": {
      "version": "v1.0.0",
      "args": ["kernel-args"]
    }
  }
}`
			return http.NewRequest(http.MethodPut, serverURL+"/api/nodes/n1", bytes.NewBuffer([]byte(testNode)))
		},
		response: `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
	},

	"read all nodes": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/nodes", nil)
		},
		response: `{"n1": {"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`,
	},

	"test idempotency (replacing existing node)": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		initFiles: []string{
			"/var/lib/warewulf/overlays/so2/rootfs/file",
			"/var/lib/warewulf/overlays/ro2/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			testNode := `{
  "node":{
    "system overlay": ["so2"],
	"runtime overlay": ["ro2"],
    "kernel": {
      "version": "v1.0.0",
      "args": ["kernel-args"]
    }
  }
}`
			return http.NewRequest(http.MethodPut, serverURL+"/api/nodes/n1", bytes.NewBuffer([]byte(testNode)))
		},
		response: `{"system overlay": ["so2"], "runtime overlay": ["ro2"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so2
    runtime overlay:
      - ro2
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
	},

	"test preventing replacing a node": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		initFiles: []string{
			"/var/lib/warewulf/overlays/so2/rootfs/file",
			"/var/lib/warewulf/overlays/ro2/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			testNode := `{
  "node":{
    "system overlay": ["so2"],
	"runtime overlay": ["ro2"],
    "kernel": {
      "version": "v1.0.0",
      "args": ["kernel-args"]
    }
  }
}`
			req, err := http.NewRequest(http.MethodPut, serverURL+"/api/nodes/n1", bytes.NewBuffer([]byte(testNode)))
			req.Header.Set("If-None-Match", "*")
			return req, err
		},
		response: `{"error": "invalid argument: node 'n1' already exists", "status": "INVALID_ARGUMENT"}`,
		status:   http.StatusBadRequest,
		resultConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
	},

	"get one specific node": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/nodes/n1", nil)
		},
		response: `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
	},

	"get one specific node with a profile": {
		initConf: `
nodeprofiles:
  default:
    system overlay:
      - so1
    runtime overlay:
      - ro1
nodes:
  n1:
    profiles:
      - default
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/nodes/n1", nil)
		},
		response: `{"profiles": ["default"], "system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
	},

	"get one specific raw node with a profile": {
		initConf: `
nodeprofiles:
  default:
    system overlay:
      - so1
    runtime overlay:
      - ro1
nodes:
  n1:
    profiles:
      - default
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/nodes/n1/raw", nil)
		},
		response: `{"profiles": ["default"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
	},

	"get unbuilt overlay info for the node": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/nodes/n1/overlays", nil)
		},
		response: `{"system overlay": { "overlays": ["so1"] }, "runtime overlay": { "overlays": ["ro1"] }}`,
	},

	"get built overlay info for the node": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		initFiles: []string{
			"/srv/warewulf/overlays/n1/__SYSTEM__.img",
			"/srv/warewulf/overlays/n1/__SYSTEM__.img.gz",
			"/srv/warewulf/overlays/n1/__RUNTIME__.img",
			"/srv/warewulf/overlays/n1/__RUNTIME__.img.gz",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/nodes/n1/overlays", nil)
		},
		response: `{
			"system overlay": {
				"overlays": ["so1"],
				"mtime": "<<PRESENCE>>"
			},
			"runtime overlay": {
				"overlays": ["ro1"],
				"mtime": "<<PRESENCE>>"
			}
		}`,
	},

	"update a node": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		request: func(serverURL string) (*http.Request, error) {
			updateNode := `{
  "node":{
    "kernel": {
      "version": "v1.0.1-newversion"
    }
  }
}`
			return http.NewRequest(http.MethodPatch, serverURL+"/api/nodes/n1", bytes.NewBuffer([]byte(updateNode)))
		},
		response: `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.1-newversion"
      args:
        - "kernel-args"
`,
	},

	"test build all nodes overlays": {
		initConf: `
nodeprofiles:
  default:
    system overlay:
      - so1
    runtime overlay:
      - ro1
nodes:
  n1:
    profiles:
      - default
  n2:
    profiles:
      - default
`,
		initFiles: []string{
			"/var/lib/warewulf/overlays/so1/rootfs/file",
			"/var/lib/warewulf/overlays/ro1/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPost, serverURL+"/api/nodes/overlays/build", nil)
		},
		response: `["n1", "n2"]`,
		resultFiles: []string{
			"/srv/warewulf/overlays/n1/__SYSTEM__.img",
			"/srv/warewulf/overlays/n1/__SYSTEM__.img.gz",
			"/srv/warewulf/overlays/n1/__RUNTIME__.img",
			"/srv/warewulf/overlays/n1/__RUNTIME__.img.gz",
			"/srv/warewulf/overlays/n2/__SYSTEM__.img",
			"/srv/warewulf/overlays/n2/__SYSTEM__.img.gz",
			"/srv/warewulf/overlays/n2/__RUNTIME__.img",
			"/srv/warewulf/overlays/n2/__RUNTIME__.img.gz",
		},
	},

	"test build one node's overlays": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		initFiles: []string{
			"/var/lib/warewulf/overlays/so1/rootfs/file",
			"/var/lib/warewulf/overlays/ro1/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPost, serverURL+"/api/nodes/n1/overlays/build", nil)
		},
		response: `"n1"`,
		resultFiles: []string{
			"/srv/warewulf/overlays/n1/__SYSTEM__.img",
			"/srv/warewulf/overlays/n1/__SYSTEM__.img.gz",
			"/srv/warewulf/overlays/n1/__RUNTIME__.img",
			"/srv/warewulf/overlays/n1/__RUNTIME__.img.gz",
		},
	},

	"test delete nodes": {
		initConf: `
nodeprofiles: {}
nodes:
  n1:
    system overlay:
      - so1
    runtime overlay:
      - ro1
    kernel:
      version: "v1.0.0"
      args:
        - "kernel-args"
`,
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/nodes/n1", nil)
		},
		response: `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`,
		resultConf: `
nodeprofiles: {}
nodes: {}
`,
	},
}

func TestNodeAPI(t *testing.T) {
	for name, tt := range nodeTests {
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

			for _, fileName := range tt.resultFiles {
				assert.FileExists(t, env.GetPath(fileName))
			}
		})
	}
}
