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

const testHwaddr = "00:00:00:00:00:01"
const testNodeName = "n1"
const testNodesConf = `
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:01
`
const testNodesConfWithAssetKey = `
nodes:
  n1:
    asset key: secret123
    network devices:
      default:
        hwaddr: 00:00:00:00:00:01
`

var filesHandlerTests = []struct {
	description string
	url         string
	body        string
	status      int
}{
	{
		"existing file",
		"/files/test.txt?wwid=" + testHwaddr,
		"hello warewulf",
		http.StatusOK,
	},
	{
		"file in subdirectory",
		"/files/subdir/test2.txt?wwid=" + testHwaddr,
		"subdir file",
		http.StatusOK,
	},
	{
		"non-existent file",
		"/files/nonexistent.txt?wwid=" + testHwaddr,
		"",
		http.StatusNotFound,
	},
	{
		"directory listing disabled at root",
		"/files/?wwid=" + testHwaddr,
		"",
		http.StatusNotFound,
	},
	{
		"directory listing disabled for subdirectory",
		"/files/subdir/?wwid=" + testHwaddr,
		"",
		http.StatusNotFound,
	},
	{
		"path traversal with ../",
		"/files/../secret.txt?wwid=" + testHwaddr,
		"",
		http.StatusNotFound,
	},
	{
		"deep path traversal",
		"/files/../../../../../../secret.txt?wwid=" + testHwaddr,
		"",
		http.StatusNotFound,
	},
	{
		"URL-encoded dot-dot traversal",
		"/files/%2e%2e/secret.txt?wwid=" + testHwaddr,
		"",
		http.StatusNotFound,
	},
	{
		"double-encoded dot-dot traversal",
		"/files/%252e%252e/secret.txt?wwid=" + testHwaddr,
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

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Paths.WWFilesdir = env.GetPath(testenv.WWFilesdir)
	conf.Warewulf.SecureP = boolPtr(false)

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

func Test_HandleFiles_NoNode(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWFilesdir+"/test.txt", "hello warewulf")

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Paths.WWFilesdir = env.GetPath(testenv.WWFilesdir)
	conf.Warewulf.SecureP = boolPtr(false)

	t.Run("no wwid and no ARP match", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/test.txt", nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("unknown wwid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/test.txt?wwid=ff:ff:ff:ff:ff:ff", nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})
}

func Test_HandleFiles_Render(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWFilesdir+"/template.ww", "node={{ .Id }}")
	env.WriteFile(testenv.WWFilesdir+"/plain.txt", "plain content")
	env.WriteFile(testenv.WWFilesdir+"/subdir/deep.ww", "hostname={{ .Hostname }}")

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Paths.WWFilesdir = env.GetPath(testenv.WWFilesdir)
	conf.Warewulf.SecureP = boolPtr(false)

	t.Run("render .ww file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/template.ww?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node="+testNodeName, string(data))
	})

	t.Run("render=nodename matching identified node succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/template.ww?render="+testNodeName+"&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node="+testNodeName, string(data))
	})

	t.Run("render=nodename mismatching identified node returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/template.ww?render=wrongnode&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("render non-.ww file returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/plain.txt?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("render nonexistent .ww file returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/missing.ww?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})

	t.Run("no render on .ww file returns raw template", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/template.ww?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node={{ .Id }}", string(data))
	})

	t.Run("render in subdirectory", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/subdir/deep.ww?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "hostname="+testNodeName, string(data))
	})

	t.Run("render without .ww suffix uses .ww fallback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/template?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node="+testNodeName, string(data))
	})

	t.Run("render without .ww suffix, neither file exists returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/missing?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})

	t.Run("path traversal with render returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/../secret.ww?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})
}

func Test_HandleFiles_AssetKey(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWFilesdir+"/test.txt", "hello warewulf")

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConfWithAssetKey)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Paths.WWFilesdir = env.GetPath(testenv.WWFilesdir)
	conf.Warewulf.SecureP = boolPtr(false)

	t.Run("assetkey required but missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/test.txt?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("assetkey required and wrong", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/test.txt?wwid="+testHwaddr+"&assetkey=wrong", nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("assetkey required and correct", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/test.txt?wwid="+testHwaddr+"&assetkey=secret123", nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "hello warewulf", string(data))
	})
}

func Test_HandleFiles_SecurePort(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWFilesdir+"/test.txt", "hello warewulf")

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Paths.WWFilesdir = env.GetPath(testenv.WWFilesdir)
	conf.Warewulf.SecureP = boolPtr(true)

	t.Run("non-privileged port rejected", func(t *testing.T) {
		// httptest.NewRequest defaults to 192.0.2.1:1234 (port >= 1024)
		req := httptest.NewRequest(http.MethodGet, "/files/test.txt?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("privileged port accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/files/test.txt?wwid="+testHwaddr, nil)
		req.RemoteAddr = "192.0.2.1:80"
		w := httptest.NewRecorder()
		HandleFiles(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "hello warewulf", string(data))
	})
}

func boolPtr(b bool) *bool {
	return &b
}
