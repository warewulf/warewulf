package warewulfd

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
)

var provisionSendTests = []struct {
	description string
	url         string
	body        string
	status      int
}{
	{"system overlay", "/overlay-system/00:00:00:ff:ff:ff", "system overlay", 200},
	{"runtime overlay", "/overlay-runtime/00:00:00:ff:ff:ff", "runtime overlay", 200},
	{"fake overlay", "/overlay-system/00:00:00:ff:ff:ff?overlay=fake", "", 404},
	{"specific overlay", "/overlay-system/00:00:00:ff:ff:ff?overlay=o1", "specific overlay", 200},
}

func Test_ProvisionSend(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "ww-test-nodes.conf-*")
	assert.NoError(t, err)
	defer file.Close()
	{
		_, err := file.WriteString(`WW_INTERNAL: 43
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:ff:ff:ff`)
		assert.NoError(t, err)
	}
	assert.NoError(t, file.Sync())
	node.ConfigFile = file.Name()
	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	provisionDir, provisionDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
	assert.NoError(t, provisionDirErr)
	defer os.RemoveAll(provisionDir)
	conf := warewulfconf.Get()
	conf.Paths.WWProvisiondir = provisionDir
	conf.Warewulf.Secure = false
	assert.NoError(t, os.MkdirAll(path.Join(provisionDir, "overlays", "n1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(provisionDir, "overlays", "n1", "__SYSTEM__.img"), []byte("system overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(provisionDir, "overlays", "n1", "__RUNTIME__.img"), []byte("runtime overlay"), 0600))
	assert.NoError(t, os.WriteFile(path.Join(provisionDir, "overlays", "n1", "o1.img"), []byte("specific overlay"), 0600))

	for _, tt := range provisionSendTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			ProvisionSend(w, req)
			res := w.Result()
			defer res.Body.Close()

			data, readErr := ioutil.ReadAll(res.Body)
			assert.NoError(t, readErr)
			assert.Equal(t, tt.body, string(data))
			assert.Equal(t, tt.status, res.StatusCode)
		})
	}
}
