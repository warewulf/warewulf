package api

import (
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

func TestOverlayAPI(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("usr/share/warewulf/overlays/testoverlay/email.ww", `
{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}
`)

	allowedNets := []net.IPNet{
		{
			IP:   net.IPv4(127, 0, 0, 0),
			Mask: net.CIDRMask(8, 32),
		},
	}
	srv := httptest.NewServer(Handler(nil, allowedNets))
	defer srv.Close()

	t.Run("get all overlays", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/overlays", nil)
		assert.NoError(t, err)

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the resp
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"testoverlay":{"files":["/email.ww"], "site":false}}`, string(body))
	})

	t.Run("get one specific overlay", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/overlays/testoverlay", nil)
		assert.NoError(t, err)

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the resp
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"files":["/email.ww"], "site":false}`, string(body))
	})

	t.Run("get overlay file", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/overlays/testoverlay/file?path=email.ww", nil)
		assert.NoError(t, err)

		// send request
		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		// validate the resp
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		// gid and uid values may vary depending on where this test is run. (local box, github, etc)
		// Assert the keys exist, but ignore the values.
		ja := jsonassert.New(t)
		ja.Assert(string(body), `
		{
			"overlay": "testoverlay",
			"path": "email.ww",
			"contents": "\n{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}\n",
			"perms": "<<PRESENCE>>",
			"uid": "<<PRESENCE>>",
			"gid": "<<PRESENCE>>"
		}`)
	})

	t.Run("create an overlay", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/overlays/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"files":null, "site":true}`, string(body))
	})

	t.Run("get all overlays", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/overlays", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"test":{"files":null, "site":true},"testoverlay":{"files":["/email.ww"], "site":false}}`, string(body))
	})

	t.Run("test delete overlays", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/overlays/test?force=true", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"files":null, "site":true}`, string(body))
	})
}
