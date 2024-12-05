package testenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func Test_Basic(t *testing.T) {
	env := New(t)
	defer env.RemoveAll(t)
	nodedb, err := node.New()
	assert.NoError(t, err)
	nodes, err := nodedb.FindAllNodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
}

func Test_two_nodes(t *testing.T) {
	env := New(t)
	env.WriteFile(t, "etc/warewulf/nodes.conf", `nodeprofiles:
  default: {}
nodes:
  node1: {}
  node2: {}
`)
	defer env.RemoveAll(t)
	nodedb, err := node.New()
	assert.NoError(t, err)
	nodes, err := nodedb.FindAllNodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
}
