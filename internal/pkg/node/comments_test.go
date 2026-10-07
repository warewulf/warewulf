package node

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_CommentPreservation(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// 1. Set up nodes.conf with a node, a profile, and various comments
	nodesConfContent := `# This is a comment at the very top of the nodes.conf file
nodeprofiles:
  # This comment is for the default profile
  default:
    # This comment is for the comment field
    comment: This profile is automatically included for each node
nodes:
  # This comment is for node1
  node1: {}
  # This is for node3
  node3: {}
`
	env.WriteFile("etc/warewulf/nodes.conf", nodesConfContent)

	// 2. Read in nodes.conf via New()
	registry, err := New()
	assert.NoError(t, err)

	// Verify original node and profile are loaded
	assert.Contains(t, registry.Nodes, "node1")
	assert.Contains(t, registry.NodeProfiles, "default")

	// 3. Manipulate the NodesYaml structure using built-in functions
	// Add a new node
	_, err = registry.AddNode("node2")
	assert.NoError(t, err)
	err = registry.DelNode("node3")
	assert.NoError(t, err)

	// Add a new profile
	_, err = registry.AddProfile("newprofile")
	assert.NoError(t, err)

	// 4. Write/Persist it back to disk
	err = registry.Persist()
	assert.NoError(t, err)

	// 5. Read the nodes.conf file back from the disk and check comments
	savedPath := warewulfconf.Get().Paths.NodesConf()
	savedData, err := os.ReadFile(savedPath)
	assert.NoError(t, err)
	savedContent := string(savedData)

	t.Logf("Saved nodes.conf content:\n%s", savedContent)

	// 6. Assert that original comments are preserved
	assert.True(t, strings.Contains(savedContent, "# This is a comment at the very top of the nodes.conf file"), "Top-level comment was lost!")
	assert.True(t, strings.Contains(savedContent, "# This comment is for the default profile"), "Default profile comment was lost!")
	assert.True(t, strings.Contains(savedContent, "# This comment is for the comment field"), "Comment field comment was lost!")
	assert.True(t, strings.Contains(savedContent, "# This comment is for node1"), "node1 comment was lost!")

	// Assert that new elements are correctly serialized
	assert.True(t, strings.Contains(savedContent, "node2"), "New node node2 was not saved!")
	assert.True(t, strings.Contains(savedContent, "newprofile"), "New profile newprofile was not saved!")

	// Assert that comment belonging to deleted element is removed
	assert.False(t, strings.Contains(savedContent, "node3"))
}
