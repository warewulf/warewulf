package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func TestImageAPI(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	authData := `
users:
- name: admin
  password hash: $2b$05$5QVWDpiWE7L4SDL9CYdi3O/l6HnbNOLoXgY2sa1bQQ7aSBKdSqvsC
`
	auth := config.NewAuthentication()
	err := auth.ParseFromRaw([]byte(authData))
	assert.NoError(t, err)

	allowedNets := []net.IPNet{
		{
			IP:   net.IPv4(127, 0, 0, 0),
			Mask: net.CIDRMask(8, 32),
		},
	}
	srv := httptest.NewServer(Handler(auth, allowedNets))
	defer srv.Close()
	env.WriteFile(path.Join(testenv.WWChrootdir, "test-image/rootfs/file"), `test`)

	t.Run("test no authentication", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/images", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.Equal(t, "Unauthorized\n", string(body))
	})

	t.Run("test get all images", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/images", nil)
		assert.NoError(t, err)
		req.SetBasicAuth("admin", "admin")

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"test-image": {"kernels":[], "size":0, "buildtime":0, "writable":true}}`, string(body))
	})

	t.Run("test get single image", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/images/test-image", nil)
		assert.NoError(t, err)
		req.SetBasicAuth("admin", "admin")

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.JSONEq(t, `{"kernels":[] ,"size":0, "buildtime":0, "writable":true}`, string(body))
	})

	t.Run("test build image", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/images/test-image/build?force=true&default=true", nil)
		assert.NoError(t, err)
		req.SetBasicAuth("admin", "admin")

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)

		var bodyData map[string]interface{}
		assert.NoError(t, json.Unmarshal([]byte(body), &bodyData))
		assert.True(t, bodyData["buildtime"].(float64) > 0.0)

		bodyData["buildtime"] = 0.0
		assert.Equal(t, map[string]interface{}{"kernels": []interface{}{}, "size": 512.0, "buildtime": 0.0, "writable": true}, bodyData)
	})

	t.Run("test rename image", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/images/test-image?build=true", bytes.NewBuffer([]byte(`{"name": "new-image"}`)))
		assert.NoError(t, err)
		req.SetBasicAuth("admin", "admin")

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)

		var bodyData map[string]interface{}
		assert.NoError(t, json.Unmarshal([]byte(body), &bodyData))
		assert.True(t, bodyData["buildtime"].(float64) > 0.0)

		bodyData["buildtime"] = 0.0
		assert.Equal(t, map[string]interface{}{"kernels": []interface{}{}, "size": 512.0, "buildtime": 0.0, "writable": true}, bodyData)
	})

	t.Run("test delete image", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/images/new-image", nil)
		assert.NoError(t, err)
		req.SetBasicAuth("admin", "admin")

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the resp
		body, err := io.ReadAll(resp.Body)
		assert.Equal(t, resp.StatusCode, http.StatusOK)
		assert.NoError(t, err)

		var bodyData map[string]interface{}
		assert.NoError(t, json.Unmarshal([]byte(body), &bodyData))
		assert.True(t, bodyData["buildtime"].(float64) > 0.0)

		bodyData["buildtime"] = 0.0
		assert.Equal(t, map[string]interface{}{"kernels": []interface{}{}, "size": 512.0, "buildtime": 0.0, "writable": true}, bodyData)
	})
}
