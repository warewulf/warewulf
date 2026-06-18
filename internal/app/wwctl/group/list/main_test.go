package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_GroupList(t *testing.T) {
	const nodesConf = `
nodeprofiles:
  gpu:
    groups:
      - rack1
nodes:
  n01:
    profiles:
      - gpu
  n02:
    groups:
      - rack1
  n03:
    groups:
      - rack1
  n04:
    groups:
      - admin
`

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "list all enumerates every referenced group",
			args: []string{},
			want: `
GROUP  MEMBERS
-----  -------
admin  n04
rack1  n01,n02,n03
`,
		},
		{
			name: "filter to a single group",
			args: []string{"rack1"},
			want: `
GROUP  MEMBERS
-----  -------
rack1  n01,n02,n03
`,
		},
		{
			name: "@all expands to every node",
			args: []string{"all"},
			want: `
GROUP  MEMBERS
-----  -------
all    n01,n02,n03,n04
`,
		},
		{
			name: "--all appends the built-in all group to the default listing",
			args: []string{"-a"},
			want: `
GROUP  MEMBERS
-----  -------
admin  n04
all    n01,n02,n03,n04
rack1  n01,n02,n03
`,
		},
		{
			name: "unknown group warns and shows empty members",
			args: []string{"missing"},
			want: `
WARN   : unknown group: missing
GROUP    MEMBERS
-----    -------
missing  --
`,
		},
		{
			name: "noheader prints comma-separated members for a single group",
			args: []string{"--noheader", "rack1"},
			want: `n01,n02,n03`,
		},
		{
			name: "noheader dedupes across multiple groups",
			args: []string{"-n", "rack1", "admin"},
			want: `n01,n02,n03,n04`,
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

func Test_GroupList_NoHeaderRequiresArg(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `
nodes:
  n01: {}
`)

	cmd := GetCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	wwlog.SetLogWriter(buf)
	cmd.SetArgs([]string{"-n"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least one group")
}
