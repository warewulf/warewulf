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
		name    string
		args    []string
		wantErr bool
		stdout  string
		inDb    string
	}{
		{
			name:    "single node list",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES   NETWORK  
  n01       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`,
		},
		{
			name:    "multiple nodes list",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]
  n02       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
`,
		},
		{
			name:    "node list returns multiple nodes",
			args:    []string{"n01,n02"},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]           
  n02       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
`,
		},
		{
			name:    "node list returns multiple nodes (case 2)",
			args:    []string{"n01,n03"},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]
  n03       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
  n03:
   profiles:
   - default
  n04:
   profiles:
   - default
  n05:
   profiles:
   - default
`,
		},
		{
			name:    "node list returns one node",
			args:    []string{"n01,"},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
`,
		},
		{
			name:    "node list profile with network",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]        default
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes:
  n01:
    profiles:
    - default
`},
	}
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		env := testenv.New(t)
		env.WriteFile(t, "etc/warewulf/nodes.conf", tt.inDb)
		var err error
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			wwlog.SetLogWriter(buf)
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err = baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Contains(t,
				strings.Join(strings.Fields(tt.stdout), ""),
				strings.Join(strings.Fields(buf.String()), ""))
		})
	}
}
