package list

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_resource_set(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		stdout  string
		inDB    string
	}{
		{
			name:    "list resource",
			args:    []string{},
			wantErr: false,
			stdout: `test1
test2
`,
			inDB: `nodeprofiles: {}
nodes: {}
resources:
  test1: {}
  test2: {}
`,
		},
		{
			name:    "list a existing resource",
			args:    []string{"-a"},
			wantErr: false,
			stdout: `test  key  value
----  ---  -----
test  foo  baar
`,
			inDB: `nodeprofiles: {}
nodes: {}
resources:
  test:
    foo: baar
`,
		},
	}
	env := testenv.New(t)
	defer env.RemoveAll()
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		env.WriteFile("etc/warewulf/nodes.conf", tt.inDB)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.stdout, buf.String())
			}
		})

	}
}
