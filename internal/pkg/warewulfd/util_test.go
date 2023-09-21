package warewulfd

import (
	"os"
	"path"
	"testing"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/stretchr/testify/assert"
)

var getOverlayFileTests = []struct {
	description string
	node        string
	context     string
	overlays    []string
	result      string
	succeed     bool
}{
	{"empty", "", "", nil, "", true},
	{"empty node", "node1", "", nil, "", true},
	{"specific overlays without node", "", "", []string{"o1", "o2"}, "overlays/o1-o2.img", true}, // will fail as node is empty
	{"system overlay", "node1", "system", []string{"o1"}, "overlays/node1/__SYSTEM__.img", true},
	{"runtime overlay", "node1", "runtime", nil, "overlays/node1/__RUNTIME__.img", true},
	{"specific overlay", "node1", "", []string{"o1"}, "overlays/node1/o1.img", true},
	{"multiple specific overlays", "node1", "", []string{"o1", "o2"}, "overlays/node1/o1-o2.img", true},
	{"multiple specific overlays with context", "node1", "system", []string{"o1", "o2"}, "overlays/node1/__SYSTEM__.img", true},
}

func Test_getOverlayFile(t *testing.T) {
	wwlog.SetLogLevel(wwlog.DEBUG)
	conf := warewulfconf.Get()
	overlayPDir, overlayPDirErr := os.MkdirTemp(os.TempDir(), "ww-test-overlay-*")
	assert.NoError(t, overlayPDirErr)
	conf.Paths.WWProvisiondir = overlayPDir
	overlayDir, overlayDirErr := os.MkdirTemp(os.TempDir(), "ww-test-provision-*")
	assert.NoError(t, overlayDirErr)
	conf.Paths.WWOverlaydir = overlayDir
	defer os.RemoveAll(overlayDir)
	assert.NoError(t, os.MkdirAll(path.Join(overlayDir, "o1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(overlayDir, "o1", "test_file_o1"), []byte("test file"), 0600))
	assert.NoError(t, os.MkdirAll(path.Join(overlayDir, "o2"), 0700))

	for _, tt := range getOverlayFileTests {
		t.Run(tt.description, func(t *testing.T) {
			var nodeInfo node.NodeInfo
			nodeInfo.Id.Set(tt.node)
			nodeInfo.RuntimeOverlay.SetSlice(tt.overlays)
			nodeInfo.SystemOverlay.SetSlice(tt.overlays)
			result, err := getOverlayFile(nodeInfo, tt.context, tt.overlays, false)
			if !tt.succeed {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.result != "" {
				tt.result = path.Join(overlayPDir, tt.result)
			}
			assert.Equal(t, tt.result, result)
		})
	}
}
