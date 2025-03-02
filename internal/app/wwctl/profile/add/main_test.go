package add

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
)

func Test_Add(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		stdout  string
		outDb   string
	}{
		{
			name:    "single profile add",
			args:    []string{"p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01: {}
nodes: {}
`,
		},
		{
			name:    "single profile add with netname and netdev",
			args:    []string{"--netname", "primary", "--netdev", "eno3", "p02"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p02:
    network devices:
      primary:
        device: eno3
nodes: {}
`,
		},
	}

	daemon.SetNoDaemon()
	for _, tt := range tests {
		env := testenv.New(t)
		env.WriteFile("etc/warewulf/nodes.conf", ``)
		var err error
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err = baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			config, configErr := node.New()
			assert.NoError(t, configErr)
			dumpBytes, _ := config.Dump()
			assert.YAMLEq(t, tt.outDb, string(dumpBytes))
		})
	}
}
