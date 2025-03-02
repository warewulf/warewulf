package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
)

func TestOverlayAPI(t *testing.T) {
	daemon.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("var/lib/warewulf/overlays/testoverlay/email.ww", `
{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}
`)

	srv := httptest.NewServer(Handler(nil))
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

		assert.JSONEq(t, `{"testoverlay":{"files":["/email.ww"]}}`, string(body))
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

		assert.JSONEq(t, `{"files":["/email.ww"]}`, string(body))
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

		assert.JSONEq(t, `{"overlay":"testoverlay","path":"email.ww","contents":"\n{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}\n"}`, string(body))
	})

	t.Run("create an overlay", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/overlays/test", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"files":null}`, string(body))
	})

	t.Run("get all overlays", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/overlays", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"test":{"files":null},"testoverlay":{"files":["/email.ww"]}}`, string(body))
	})

	t.Run("test delete overlays", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/overlays/test?force=true", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NoError(t, resp.Body.Close())

		assert.JSONEq(t, `{"files":null}`, string(body))
	})
}
