package imprt

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Node_Import(t *testing.T) {
	tests := map[string]struct {
		args       []string
		importFile string
		wantErr    bool
		inDB       string
		outDB      string
	}{
		"import new node": {
			args: []string{"importFile"},
			importFile: `
n1:
  id: n1
  profiles:
  - default
  network devices:
    eth0:
      device: eth0
      hwaddr: c4:cb:e1:bb:dd:e9
      ipaddr: 192.168.1.10`,
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes: {}`,
			outDB: `
nodeprofiles: {}
nodes:
  n1:
    profiles:
    - default
    network devices:
      eth0:
        device: eth0
        hwaddr: c4:cb:e1:bb:dd:e9
        ipaddr: 192.168.1.10`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			{
				wd, err := os.Getwd()
				assert.NoError(t, err)
				defer func() { assert.NoError(t, os.Chdir(wd)) }()
			}
			assert.NoError(t, os.Chdir(env.GetPath(".")))
			env.WriteFile("./importFile", tt.importFile)
			env.WriteFile("etc/warewulf/nodes.conf", tt.inDB)
			warewulfd.SetNoDaemon()

			baseCmd := GetCommand()
			args := append(tt.args, "--yes")
			baseCmd.SetArgs(args)
			buf := new(bytes.Buffer)
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
