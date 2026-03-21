package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_List(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stdout string
		inDb   string
	}{
		{
			name: "image list test",
			args: []string{"-s"},
			stdout: `
IMAGE NAME  NODES  SIZE
----------  -----  ----
test        1      0 B
`,
			inDb: `
nodeprofiles:
  default: {}
nodes:
  n01:
    image name: test
    profiles:
    - default
`,
		},
	}

	for _, tt := range tests {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.WriteFile("etc/warewulf/nodes.conf", tt.inDb)
		env.MkdirAll("var/lib/warewulf/chroots/test/rootfs")

		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.stdout), strings.TrimSpace(buf.String()))
		})
	}
}
