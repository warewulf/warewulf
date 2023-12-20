package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
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
			name: "profile list test",
			args: []string{},
			stdout: `PROFILE NAME  COMMENT/DESCRIPTION
  default`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`},
		{
			name: "profile list returns multiple profiles",
			args: []string{"default,test"},
			stdout: `PROFILE NAME  COMMENT/DESCRIPTION
			  default
			  test`,
			inDb: `WW_INTERNAL: 43
nodeprofiles:
  default: {}
  test: {}
nodes:
  n01:
    profiles:
    - default
`,
		}, /*
					{
						name: "profile list returns one profiles",
						args: []string{"test,"},
						stdout: `PROFILE NAME  COMMENT/DESCRIPTION
			  test          --`,
						inDb: `WW_INTERNAL: 43
			nodeprofiles:
			  default: {}
			  test: {}
			nodes:
			  n01:
			    profiles:
			    - default
			`,
					},
					{
						name: "profile list returns all profiles",
						args: []string{","},
						stdout: `PROFILE NAME  COMMENT/DESCRIPTION
			  default       --
			  test          --`,
						inDb: `WW_INTERNAL: 43
			nodeprofiles:
			  default: {}
			  test: {}
			nodes:
			  n01:
			    profiles:
			    - default
			`,
					},*/
	}

	warewulfd.SetNoDaemon()
	//wwlog.SetLogLevel(wwlog.DEBUG)
	for _, tt := range tests {
		env := testenv.New(t)
		env.WriteFile(t, "etc/warewulf/nodes.conf",
			tt.inDb)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			assert.NoError(t, baseCmd.Execute())
			assert.Equal(t,
				strings.Join(strings.Fields(tt.stdout), ""),
				strings.Join(strings.Fields(buf.String()), ""))
		})
	}
}
