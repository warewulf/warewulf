// SPDX-License-Identifier: BSD-3-Clause
// Copyright holders are listed in LICENSE.md.

package warewulfd

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func TestBootTemplateHandlers(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `nodes:
  n1:
    asset key: test-key
    kernel:
      version: test-version
    tags:
      label: node
    network devices:
      default:
        hwaddr: 02:00:00:00:00:01`)
	require.NoError(t, LoadNodeDB())

	stages := []struct {
		name     string
		template string
		handler  http.HandlerFunc
	}{
		{"ipxe", "etc/warewulf/ipxe/default.ipxe", HandleIpxe},
		{"grub", "etc/warewulf/grub/grub.cfg.ww", HandleGrub},
	}
	const key = "?assetkey=test-key"
	tests := []struct {
		name   string
		source string
		query  string
		status int
		body   string
	}{
		{"render", "{{.Id}} {{upper .Tags.label}} {{.Tags.missing}}\n", key, http.StatusOK, "n1 NODE <no value>\n"},
		{"parse error", "{{if}}", key, http.StatusInternalServerError, ""},
		{"execution error after output", "partial output{{.MissingField}}", key, http.StatusInternalServerError, ""},
		// Unreachable branches check function availability without invoking the functions.
		{"no overlay input function", `{{if false}}{{Include "unused"}}{{end}}`, key, http.StatusInternalServerError, ""},
		{"no overlay output function", `{{if false}}{{file "unused"}}{{end}}`, key, http.StatusInternalServerError, ""},
		{"missing asset key", "script", "", http.StatusUnauthorized, ""},
		{"compressed template", "script", key + "&compress=gz", http.StatusNotFound, ""},
	}
	for _, stage := range stages {
		for _, tt := range tests {
			t.Run(stage.name+"/"+tt.name, func(t *testing.T) {
				env.WriteFile(stage.template, tt.source)
				req := httptest.NewRequest(http.MethodGet, "/"+stage.name+"/02:00:00:00:00:01"+tt.query, nil)
				w := httptest.NewRecorder()
				stage.handler(w, req)
				resp := w.Result()
				defer func() { _ = resp.Body.Close() }()

				assert.Equal(t, tt.status, resp.StatusCode)
				assert.Equal(t, tt.body, w.Body.String())
				if tt.status == http.StatusOK {
					assert.Equal(t, "text", resp.Header.Get("Content-Type"))
					assert.Equal(t, strconv.Itoa(len(tt.body)), resp.Header.Get("Content-Length"))
				}
			})
		}
	}
}
