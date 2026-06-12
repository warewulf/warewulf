package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_NodegroupList(t *testing.T) {
	const nodesConf = `
nodeprofiles:
  gpu:
    nodegroups:
      - rack1
nodes:
  n01:
    profiles:
      - gpu
  n02:
    nodegroups:
      - rack1
  n03: {}
  n04: {}
nodegroups:
  rack1:
    - n03
  admin:
    - n04
`

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "list all enumerates every defined nodegroup",
			args: []string{},
			want: `
NODEGROUP  MEMBERS
---------  -------
admin      n04
rack1      n01,n02,n03
`,
		},
		{
			name: "filter to a single nodegroup",
			args: []string{"rack1"},
			want: `
NODEGROUP  MEMBERS
---------  -------
rack1      n01,n02,n03
`,
		},
		{
			name: "@all expands to every node",
			args: []string{"all"},
			want: `
NODEGROUP  MEMBERS
---------  -------
all        n01,n02,n03,n04
`,
		},
		{
			name: "unknown nodegroup warns and shows empty members",
			args: []string{"missing"},
			want: `
WARN   : unknown nodegroup: missing
NODEGROUP  MEMBERS
---------  -------
missing    --
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("etc/warewulf/nodes.conf", nodesConf)

			cmd := GetCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			cmd.SetArgs(tt.args)
			assert.NoError(t, cmd.Execute())
			assert.Equal(t, strings.TrimSpace(tt.want), strings.TrimSpace(buf.String()))
		})
	}
}
