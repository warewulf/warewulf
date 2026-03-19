package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var filesHandlerTests = []struct {
	description string
	url         string
	body        string
	status      int
}{
	{
		"existing file",
		"/files/test.txt",
		"hello warewulf",
		http.StatusOK,
	},
	{
		"file in subdirectory",
		"/files/subdir/test2.txt",
		"subdir file",
		http.StatusOK,
	},
	{
		"non-existent file",
		"/files/nonexistent.txt",
		"",
		http.StatusNotFound,
	},
	{
		"directory listing disabled at root",
		"/files/",
		"",
		http.StatusNotFound,
	},
	{
		"directory listing disabled for subdirectory",
		"/files/subdir/",
		"",
		http.StatusNotFound,
	},
	{
		"path traversal with ../",
		"/files/../secret.txt",
		"",
		http.StatusNotFound,
	},
	{
		"deep path traversal",
		"/files/../../../../../../secret.txt",
		"",
		http.StatusNotFound,
	},
	{
		"URL-encoded dot-dot traversal",
		"/files/%2e%2e/secret.txt",
		"",
		http.StatusNotFound,
	},
	{
		"double-encoded dot-dot traversal",
		"/files/%252e%252e/secret.txt",
		"",
		http.StatusNotFound,
	},
}

func Test_HandleFiles(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWFilesdir+"/test.txt", "hello warewulf")
	env.WriteFile(testenv.WWFilesdir+"/subdir/test2.txt", "subdir file")
	// Sentinel file outside the files dir: if traversal succeeds it would be
	// served as 200; keeping it here makes traversal failures detectable.
	env.WriteFile("secret.txt", "secret content")

	conf := warewulfconf.Get()
	conf.Paths.WWFilesdir = env.GetPath(testenv.WWFilesdir)

	for _, tt := range filesHandlerTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			HandleFiles(w, req)
			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			assert.Equal(t, tt.status, res.StatusCode)
			if tt.body != "" {
				data, readErr := io.ReadAll(res.Body)
				assert.NoError(t, readErr)
				assert.Equal(t, tt.body, string(data))
			}
		})
	}
}
