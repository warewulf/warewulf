package ww4test

import (
	"os"
	"testing"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/stretchr/testify/assert"
)

func Test_Basic(t *testing.T) {
	Env.New(t)
	defer os.RemoveAll(Env.BaseDir)
	nodedb, err := node.New()
	assert.NoError(t, err)
	nodes, err := nodedb.FindAllNodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
}

func Test_two_nodes(t *testing.T) {
	Env.NodesConf = `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  node1: {}
  node2: {}
`
	Env.New(t)
	defer os.RemoveAll(Env.BaseDir)
	nodedb, err := node.New()
	assert.NoError(t, err)
	nodes, err := nodedb.FindAllNodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
}
