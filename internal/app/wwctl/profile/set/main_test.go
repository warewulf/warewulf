package set

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
)

type test_description struct {
	name    string
	args    []string
	wantErr bool
	stdout  string
	inDB    string
	outDb   string
}

func run_test(t *testing.T, test test_description) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", test.inDB)
	daemon.SetNoDaemon()
	name := test.name
	if name == "" {
		name = t.Name()
	}
	t.Run(name, func(t *testing.T) {
		baseCmd := GetCommand()
		test.args = append(test.args, "--yes")
		baseCmd.SetArgs(test.args)
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		err := baseCmd.Execute()
		if test.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, buf.String(), test.stdout)
			content := env.ReadFile("etc/warewulf/nodes.conf")
			assert.YAMLEq(t, test.outDb, content)
		}
	})
}

func Test_Set_Netdev(t *testing.T) {
	test := test_description{
		args:    []string{"--netname=default", "--netdev=eth0", "default"},
		wantErr: false,
		stdout:  "",
		inDB: `nodeprofiles:
  default: {}
nodes: {}
`,
		outDb: `nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes: {}
`}
	run_test(t, test)
}
func Test_Set_Netdev_and_Mask(t *testing.T) {
	test := test_description{
		args:    []string{"--netname=default", "--netdev=eth0", "-M=255.255.255.0", "default"},
		wantErr: false,
		stdout:  "",
		inDB: `nodeprofiles:
  default: {}
nodes: {}
`,
		outDb: `nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        netmask: 255.255.255.0
nodes: {}
`}
	run_test(t, test)
}

func Test_Set_Mask_Existing_NetDev(t *testing.T) {
	test := test_description{
		args:    []string{"--netname=default", "-M=255.255.255.0", "default"},
		wantErr: false,
		stdout:  "",
		inDB: `nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes: {}
`,
		outDb: `nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        netmask: 255.255.255.0
nodes: {}
`}
	run_test(t, test)
}
