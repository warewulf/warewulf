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
	result      string
}{
	{
		description: "empty inputs produces no result",
		node:        "",
		context:     "",
		result:      "",
	},
	{
		description: "a node with no context produces no result",
		node:        "node1",
		context:     "",
		result:      "",
	},
	{
		description: "system overlay for a node points to the node's system overlay image",
		node:        "node1",
		context:     "system",
		result:      "node1/__SYSTEM__.img",
	},
	{
		description: "runtime overlay for a node points to the node's runtime overlay image",
		node:        "node1",
		context:     "runtime",
		result:      "node1/__RUNTIME__.img",
	},
}

func Test_getOverlayFile(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile("etc/warewulf/nodes.conf", `
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
			result, err := getOverlayFile(nodeInfo, tt.context, false)
			assert.NoError(t, err)
			if tt.result != "" {
				tt.result = path.Join(conf.Paths.OverlayProvisiondir(), tt.result)
			}
			assert.Equal(t, tt.result, result)
		})
	}
}
