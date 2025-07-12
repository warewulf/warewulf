package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func TestNodeAPI(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
	env.MkdirAll("/var/lib/warewulf/overlays/so1")
	env.MkdirAll("/var/lib/warewulf/overlays/ro1")
	defer env.RemoveAll()

	allowedNets := []net.IPNet{
		{
			IP:   net.IPv4(127, 0, 0, 0),
			Mask: net.CIDRMask(8, 32),
		},
	}
	srv := httptest.NewServer(Handler(nil, allowedNets))
	defer srv.Close()

	t.Run("add a node", func(t *testing.T) {
		// prepareration

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
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/nodes/test", bytes.NewBuffer([]byte(testNode)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("read all nodes", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes", nil)
		assert.NoError(t, err)

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the resp
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"node1": {}, "test": {"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`, string(body))
	})

	t.Run("get one specific node", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes/test", nil)
		assert.NoError(t, err)

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the resp
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("get unbuilt overlay info for the node", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes/test/overlays", nil)
		assert.NoError(t, err)

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the response
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		ja := jsonassert.New(t)
		ja.Assertf(string(body), `{
			"system overlay": {
				"overlays": ["so1"]
			},
			"runtime overlay": {
				"overlays": ["ro1"]
			}
		}`)
	})

	t.Run("update the node", func(t *testing.T) {
		updateNode := `{
  "node":{
    "kernel": {
      "version": "v1.0.1-newversion"
    }
  }
}`
		req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/nodes/test", bytes.NewBuffer([]byte(updateNode)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("get one specific node (again)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("get one specific (raw) node", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes/test/raw", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("test build all nodes overlays", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/nodes/overlays/build", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `["node1", "test"]`, string(body))
	})

	t.Run("test build one node's overlays", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/nodes/test/overlays/build", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `"test"`, string(body))
	})

	t.Run("get built overlay info for the node", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes/test/overlays", nil)
		assert.NoError(t, err)

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the response
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		ja := jsonassert.New(t)
		ja.Assertf(string(body), `{
			"system overlay": {
				"overlays": ["so1"],
				"mtime": "<<PRESENCE>>"
			},
			"runtime overlay": {
				"overlays": ["ro1"],
				"mtime": "<<PRESENCE>>"
			}
		}`)

		data := map[string]any{}
		assert.NoError(t, json.Unmarshal(body, &data))
		{
			_, err := time.Parse(time.RFC3339, data["system overlay"].(map[string]any)["mtime"].(string))
			assert.NoError(t, err)
		}
		{
			_, err := time.Parse(time.RFC3339, data["runtime overlay"].(map[string]any)["mtime"].(string))
			assert.NoError(t, err)
		}
	})

	t.Run("test delete nodes", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/nodes/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"system overlay": ["so1"], "runtime overlay": ["ro1"], "kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})
}
