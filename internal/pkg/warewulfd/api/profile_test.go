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

func TestProfileAPI(t *testing.T) {
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

	t.Run("get all profiles", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/profiles", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"default": {}}`, string(body))
	})

	t.Run("add a new profile", func(t *testing.T) {
		testProfile := `{"profile": {"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/profiles/test", bytes.NewBuffer([]byte(testProfile)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`, string(body))

		// Add the same node again. It should fail since the name is duplicated.
		req, err = http.NewRequest(http.MethodPut, srv.URL+"/api/profiles/test", bytes.NewBuffer([]byte(testProfile)))
		assert.NoError(t, err)

		resp, err = http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		assert.Equal(t, 400, resp.StatusCode) // Invalid aurgument error
	})

	t.Run("re-read all profiles", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/profiles", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"default": {}, "test": {"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}}`, string(body))
	})

	t.Run("get one specific profile (that was just added)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/profiles/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"kernel": {"version": "v1.0.0", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("update the profile", func(t *testing.T) {
		updateProfile := `{"profile": {"kernel": {"version": "v1.0.1-newversion"}}}`
		req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/profiles/test", bytes.NewBuffer([]byte(updateProfile)))
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("get one specific profile (that was just updated)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/profiles/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})

	t.Run("test delete a profile", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/profiles/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"kernel": {"version": "v1.0.1-newversion", "args": ["kernel-args"]}}`, string(body))
	})
}
