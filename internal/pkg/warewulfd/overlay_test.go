package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var systemOverlayTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"system overlay", "/system/00:00:00:ff:ff:ff", "system overlay", 200, "10.10.10.10:9873"},
}

var runtimeOverlayTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{"runtime overlay", "/runtime/00:00:00:ff:ff:ff", "runtime overlay", 200, "10.10.10.10:9873"},
}

func Test_HandleSystemRuntimeOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("/etc/warewulf/nodes.conf", `nodeprofiles:
  default:
    image name: suse
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:ff:ff:ff
    profiles:
    - default`)

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse

	assert.NoError(t, os.MkdirAll(path.Join(conf.Paths.OverlayProvisiondir(), "n1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__SYSTEM__.img"), []byte("system overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.OverlayProvisiondir(), "n1", "__RUNTIME__.img"), []byte("runtime overlay"), 0600))

	for _, tt := range systemOverlayTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleSystemOverlay(w, req)
			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			data, readErr := io.ReadAll(res.Body)
			assert.NoError(t, readErr)
			if tt.body != "" {
				assert.Equal(t, tt.body, string(data))
			}
			assert.Equal(t, tt.status, res.StatusCode)
		})
	}

	for _, tt := range runtimeOverlayTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleRuntimeOverlay(w, req)
			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			data, readErr := io.ReadAll(res.Body)
			assert.NoError(t, readErr)
			if tt.body != "" {
				assert.Equal(t, tt.body, string(data))
			}
			assert.Equal(t, tt.status, res.StatusCode)
		})
	}
}

func Test_HandleOverlayFile(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWOverlaydir+"/testoverlay/rootfs/etc/plain.conf", "plain content")
	env.WriteFile(testenv.WWOverlaydir+"/testoverlay/rootfs/etc/template.ww", "node={{ .Id }}")
	env.WriteFile(testenv.WWOverlaydir+"/testoverlay/rootfs/etc/notww.txt", "not a template")

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Warewulf.SecureP = boolPtr(false)

	t.Run("raw non-.ww file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/plain.conf?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "plain content", string(data))
	})

	t.Run("raw .ww file without render", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/template.ww?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node={{ .Id }}", string(data))
	})

	t.Run("render .ww file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/template.ww?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node="+testNodeName, string(data))
	})

	t.Run("render without .ww suffix uses .ww fallback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/template?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node="+testNodeName, string(data))
	})

	t.Run("render=nodename matching identified node succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/template.ww?render="+testNodeName+"&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "node="+testNodeName, string(data))
	})

	t.Run("render=nodename mismatching identified node returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/template.ww?render=wrongnode&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("render non-.ww file returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/notww.txt?render&wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("nonexistent overlay returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/nosuchoverlay/etc/file.conf?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})

	t.Run("nonexistent file in overlay returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/missing.conf?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})

	t.Run("no node identification returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/plain.conf", nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("path traversal returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/../../../etc/passwd?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	})

	t.Run("no overlay specified returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file//etc/plain.conf?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("no path specified returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})
}

func Test_HandleOverlayFile_AssetKey(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWOverlaydir+"/testoverlay/rootfs/etc/plain.conf", "plain content")

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConfWithAssetKey)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Warewulf.SecureP = boolPtr(false)

	t.Run("assetkey required but missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/plain.conf?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("assetkey required and wrong", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/plain.conf?wwid="+testHwaddr+"&assetkey=wrong", nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("assetkey required and correct", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/plain.conf?wwid="+testHwaddr+"&assetkey=secret123", nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "plain content", string(data))
	})
}

func Test_HandleOverlayFile_SecurePort(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile(testenv.WWOverlaydir+"/testoverlay/rootfs/etc/plain.conf", "plain content")

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	assert.NoError(t, LoadNodeDB())

	conf := warewulfconf.Get()
	conf.Warewulf.SecureP = boolPtr(true)

	t.Run("non-privileged port rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/plain.conf?wwid="+testHwaddr, nil)
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("privileged port accepted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/overlay-file/testoverlay/etc/plain.conf?wwid="+testHwaddr, nil)
		req.RemoteAddr = "192.0.2.1:80"
		w := httptest.NewRecorder()
		HandleOverlayFile(w, req)
		res := w.Result()
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "plain content", string(data))
	})
}
