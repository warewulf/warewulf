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

func TestPowerAPI(t *testing.T) {
	// same as TestNodeAPI but for power API.
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

	t.Run("add a node with no ipmi", func(t *testing.T) {

		testNode := `{
  "node":{
    "kernel": {
      "version": "v1.0.0",
      "args": ["kernel-args"]
    }
  }
}`
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/nodes/testing1234OverOver", bytes.NewBuffer([]byte(testNode)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`, string(body))
	})

	// add a new profile node000profile profile.
	t.Run("add test testing1234OverOver profile", func(t *testing.T) {
		testProfile := `{"profile": {"ipmi": {"template": "ipmitool.tmpl"}}}`
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/profiles/testing1234OverOverprofile", bytes.NewBuffer([]byte(testProfile)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"ipmi": {"template": "ipmitool.tmpl"}}`, string(body))
	})

	//add a node with ipmi config but no BMC connection.
	t.Run("add test ipmi node", func(t *testing.T) {
		ipmiTestNodeConfig :=
			`{
		"node":{
			"ipmi": {"username": "ADMIN", "password": "ADMIN", "ipaddr": "127.0.0.1", "template": "ipmitool.tmpl","profiles": ["default"]}
		}
		}`
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/nodes/ipminode", bytes.NewBuffer([]byte(ipmiTestNodeConfig)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"ipmi": {"username": "ADMIN", "password": "ADMIN", "ipaddr": "127.0.0.1", "template": "ipmitool.tmpl"}}`, string(body))

		req, err = http.NewRequest(http.MethodGet, srv.URL+"/api/power/ipminode", nil)
		assert.NoError(t, err)

		resp, err = http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		_, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
	})

	t.Run("get node power, ipmi not configured", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/power/testing1234OverOver", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"error": "invalid argument: node ipmi not configured", "status":"INVALID_ARGUMENT"}`, string(body))
	})

	t.Run("set node power on, ipmi not configured", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/power/testing1234OverOver", bytes.NewBuffer([]byte(`{"state": "on"}`)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"error": "invalid argument: node ipmi not configured", "status":"INVALID_ARGUMENT"}`, string(body))
	})

	t.Run("set node power off, ipmi not configured", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/power/testing1234OverOver", bytes.NewBuffer([]byte(`{"state": "off"}`)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"error": "invalid argument: node ipmi not configured", "status":"INVALID_ARGUMENT"}`, string(body))
	})

	t.Run("set node power cycle, ipmi not configured", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/power/testing1234OverOver", bytes.NewBuffer([]byte(`{"state": "cycle"}`)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"error": "invalid argument: node ipmi not configured", "status":"INVALID_ARGUMENT"}`, string(body))
	})

	t.Run("set node power soft, ipmi not configured", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/power/testing1234OverOver", bytes.NewBuffer([]byte(`{"state": "soft"}`)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"error": "invalid argument: node ipmi not configured", "status":"INVALID_ARGUMENT"}`, string(body))
	})

	t.Run("set node power reset, ipmi not configured", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/power/testing1234OverOver", bytes.NewBuffer([]byte(`{"state": "reset"}`)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"error": "invalid argument: node ipmi not configured", "status":"INVALID_ARGUMENT"}`, string(body))
	})

	t.Run("set node power invalid, ipmi not configured", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/power/testing1234OverOver", bytes.NewBuffer([]byte(`{"state": "invalid"}`)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"error": "invalid argument: node ipmi not configured", "status":"INVALID_ARGUMENT"}`, string(body))
	})

}
