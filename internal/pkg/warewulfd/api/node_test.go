package api

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func TestNodeAPI(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
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

		assert.JSONEq(t, `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`, string(body))
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

		assert.JSONEq(t, `{"node1": {}, "test": {"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`, string(body))
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

		assert.JSONEq(t, `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`, string(body))
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

		assert.JSONEq(t, `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("get one specific node (again)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("get one specific (raw) node", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/nodes/test/raw", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
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

	t.Run("test delete nodes", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/nodes/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})
}
