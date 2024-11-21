package set

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_resource_set(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		stdout  string
		inDB    string
		outDb   string
	}{
		{
			name:    "set resource",
			args:    []string{"--tagset", "foo=test", "test"},
			wantErr: false,
			stdout:  "",
			inDB: `nodeprofiles: {}
nodes: {}
resource:
  test:
    foo: baar
`,
			outDb: `nodeprofiles: {}
nodes: {}
resource:
  test:
    foo: test
`},
		{
			name:    "set non exiting resource",
			args:    []string{"--tagset", "foo=test", "none"},
			wantErr: true,
			stdout:  "",
			inDB: `nodeprofiles: {}
nodes: {}
resource:
  test:
    foo: baar
`,
			outDb: ``},
	}
	env := testenv.New(t)
	defer env.RemoveAll(t)
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		env.WriteFile(t, "etc/warewulf/nodes.conf", tt.inDB)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err := baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, buf.String(), tt.stdout)
				content := env.ReadFile(t, "etc/warewulf/nodes.conf")
				assert.YAMLEq(t, tt.outDb, content)
			}
		})

	}
}
