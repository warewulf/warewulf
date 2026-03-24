package delete

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Delete(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		inDB    string
		outDB   string
		wantErr bool
	}{
		{
			name: "delete single node",
			args: []string{"--yes", "n01"},
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default: {}
nodes:
  n02:
    profiles:
    - default`,
		},
		{
			name: "delete multiple nodes",
			args: []string{"--yes", "n01", "n02"},
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
    profiles:
    - default
  n03:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default: {}
nodes:
  n03:
    profiles:
    - default`,
		},
		{
			name: "delete non-existent node",
			args: []string{"--yes", "doesnotexist"},
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default`,
		},
		{
			name: "delete all nodes",
			args: []string{"--yes", "n01", "n02"},
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01: {}
  n02: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
	}

	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("etc/warewulf/nodes.conf", tt.inDB)

			buf := new(bytes.Buffer)
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err := baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				content := env.ReadFile("etc/warewulf/nodes.conf")
				assert.YAMLEq(t, tt.outDB, content)
			}
		})
	}
}
