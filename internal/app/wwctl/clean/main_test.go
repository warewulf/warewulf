package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Clean(t *testing.T) {
	wwlog.SetLogLevel(wwlog.DEBUG)
	env := testenv.New(t)
	env.WriteFile(t, node.GetNodesConf("etc"),
		`nodeprofiles: {}
nodes:
  node1: {}
`)
	env.WriteFile(t, "srv/warewulf/overlays/node1/__SYSTEM__.img", "Fake System")
	env.WriteFile(t, "srv/warewulf/overlays/node2/__SYSTEM__.img", "Fake System")
	env.WriteFile(t, "var/cache/warewulf/test", "Nothing to see here")
	baseCmd := GetCommand()
	err := baseCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, env.GetPath("srv/warewulf/overlays/node1/__SYSTEM__.img"))
	assert.NoFileExists(t, env.GetPath("srv/warewulf/overlays/node2/__SYSTEM__.img"))
	assert.NoFileExists(t, env.GetPath("/var/cache/warewulf/test"))
}
