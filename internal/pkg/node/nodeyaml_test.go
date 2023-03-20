package node


import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)


func writeTestConfigFile(data string) (f *os.File, err error) {
	f, err = ioutil.TempFile("", "nodes.conf-*")
	if err != nil {
		return f, err
	} else {
		_, err = f.WriteString(data)
		if err != nil {
			return f, err
		} else {
			err = f.Sync()
			return f, err
		}
	}
}


func Test_ReadNodeYamlFromFileMinimal(t *testing.T) {
	file, writeErr := writeTestConfigFile(`
nodeprofiles:
  default:
    comment: A default profile
nodes:
  test_node:
    comment: A single node`)
	if file != nil {
		defer os.Remove(file.Name())
	}
	assert.NoError(t, writeErr)

	nodeYaml, err := ReadNodeYamlFromFile(file.Name())
	assert.NoError(t, err)
	assert.Contains(t, nodeYaml.NodeProfiles, "default")
	assert.Equal(t, nodeYaml.NodeProfiles["default"].Comment, "A default profile")
	assert.Contains(t, nodeYaml.Nodes, "test_node")
	assert.Equal(t, nodeYaml.Nodes["test_node"].Comment, "A single node")
}
