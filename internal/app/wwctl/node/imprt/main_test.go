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
			args: []string{"--csv=false"},
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
		"import from CSV": {
			args: []string{"--csv=true"},
			importFile: `nodename,net.default.hwaddr,net.default.ipaddr,net.default.netmask,net.default.gateway,net.default.netdev,discoverable,image
n1,11:22:33:44:55:66,192.168.1.10,255.255.255.0,192.168.1.1,eth0,false,rocky-9`,
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes: {}`,
			outDB: `
nodeprofiles: {}
nodes:
  n1:
    image name: rocky-9
    discoverable: "false"
    network devices:
      default:
        device: eth0
        hwaddr: 11:22:33:44:55:66
        ipaddr: 192.168.1.10
        netmask: 255.255.255.0
        gateway: 192.168.1.1`,
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
			args := append(tt.args, "importFile", "--yes")
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
