package warewulfd

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
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
		result:      "node1/o1-o2.img",
	},
	{
		description: "system overlay for a node points to the node's system overlay image",
		node:        "node1",
		context:     "system",
		overlays:    []string{"o1"},
		result:      "node1/__SYSTEM__.img",
	},
	{
		description: "runtime overlay for a node points to the node's runtime overlay image",
		node:        "node1",
		context:     "runtime",
		overlays:    nil,
		result:      "node1/__RUNTIME__.img",
	},
	{
		description: "a specific overlay for a node points to that specific overlay image for that node",
		node:        "node1",
		context:     "",
		overlays:    []string{"o1"},
		result:      "node1/o1.img",
	},
	{
		description: "a specific set of overlays for a node points to a combined overlay image for that node",
		node:        "node1",
		context:     "",
		overlays:    []string{"o1", "o2"},
		result:      "node1/o1-o2.img",
	},
	{
		description: "a specific set of overlays for a node while also specifying a context points to the contextual overlay image for that node",
		node:        "node1",
		context:     "system",
		overlays:    []string{"o1", "o2"},
		result:      "node1/__SYSTEM__.img",
	},
}

func Test_getOverlayFile(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile(t, node.GetNodesConf("etc"), `
nodes:
  node1: {} `)
	conf := warewulfconf.Get()
	assert.NoError(t, os.MkdirAll(path.Join(conf.Paths.WWOverlaydir, "o1"), 0700))
	assert.NoError(t, os.WriteFile(path.Join(conf.Paths.WWOverlaydir, "o1", "test_file_o1"), []byte("test file"), 0600))
	assert.NoError(t, os.MkdirAll(path.Join(conf.Paths.WWOverlaydir, "o2"), 0700))
	nodeDB, err := node.New()
	assert.NoError(t, err)
	for _, tt := range getOverlayFileTests {
		t.Run(tt.description, func(t *testing.T) {
			nodeInfo, err := nodeDB.GetNode("node1")
			assert.NoError(t, err)
			nodeInfo.RuntimeOverlay = tt.overlays
			nodeInfo.SystemOverlay = tt.overlays
			result, err := getOverlayFile(nodeInfo, tt.context, tt.overlays, false)
			assert.NoError(t, err)
			if tt.result != "" {
				tt.result = path.Join(conf.Paths.OverlayProvisiondir(), tt.result)
			}
			assert.Equal(t, tt.result, result)
		})
	}
}
