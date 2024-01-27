package warewulfd

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var getOverlayFileTests = []struct {
	description string
	node        string
	context     string
	overlays    []string
	result      string
}{
	{
		description: "empty inputs produces no result",
		node:        "",
		context:     "",
		overlays:    nil,
		result:      "",
	},
	{
		description: "a node with no context or overlays produces no result",
		node:        "node1",
		context:     "",
		overlays:    nil,
		result:      "",
	},
	{
		description: "overlays with no node or context points to a combined overlay image",
		node:        "",
		context:     "",
		overlays:    []string{"o1", "o2"},
		result:      "overlays/o1-o2.img",
	},
	{
		description: "system overlay for a node points to the node's system overlay image",
		node:        "node1",
		context:     "system",
		overlays:    []string{"o1"},
		result:      "overlays/node1/__SYSTEM__.img",
	},
	{
		description: "runtime overlay for a node points to the node's runtime overlay image",
		node:        "node1",
		context:     "runtime",
		overlays:    nil,
		result:      "overlays/node1/__RUNTIME__.img",
	},
	{
		description: "a specific overlay for a node points to that specific overlay image for that node",
		node:        "node1",
		context:     "",
		overlays:    []string{"o1"},
		result:      "overlays/node1/o1.img",
	},
	{
		description: "a specific set of overlays for a node points to a combined overlay image for that node",
		node:        "node1",
		context:     "",
		overlays:    []string{"o1", "o2"},
		result:      "overlays/node1/o1-o2.img",
	},
	{
		description: "a specific set of overlays for a node while also specifying a context points to the contextual overlay image for that node",
		node:        "node1",
		context:     "system",
		overlays:    []string{"o1", "o2"},
		result:      "overlays/node1/__SYSTEM__.img",
	},
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
			assert.NoError(t, err)
			if tt.result != "" {
				tt.result = path.Join(overlayPDir, tt.result)
			}
			assert.Equal(t, tt.result, result)
		})
	}
}
