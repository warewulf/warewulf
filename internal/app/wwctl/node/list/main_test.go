package list

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
			stdout: `  NODE NAME  PROFILES  NETWORK  
  n01        default            
`,
			inDb: `WW_INTERNAL: 43
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
  n01        default            
  n02        default
`,
			inDb: `WW_INTERNAL: 43
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
  n01        default            
  n02        default
`,
			inDb: `WW_INTERNAL: 43
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
  n01        default            
  n03        default
`,
			inDb: `WW_INTERNAL: 43
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
  n01        default            
`,
			inDb: `WW_INTERNAL: 43
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
	}
	conf_yml := `WW_INTERNAL: 0`
	tempWarewulfConf, warewulfConfErr := os.CreateTemp("", "warewulf.conf-")
	assert.NoError(t, warewulfConfErr)
	defer os.Remove(tempWarewulfConf.Name())
	_, warewulfConfErr = tempWarewulfConf.Write([]byte(conf_yml))
	assert.NoError(t, warewulfConfErr)
	assert.NoError(t, tempWarewulfConf.Sync())
	assert.NoError(t, warewulfconf.New().Read(tempWarewulfConf.Name()))

	tempNodeConf, nodesConfErr := os.CreateTemp("", "nodes.conf-")
	assert.NoError(t, nodesConfErr)
	defer os.Remove(tempNodeConf.Name())
	node.ConfigFile = tempNodeConf.Name()
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		var err error
		_, err = tempNodeConf.Seek(0, 0)
		assert.NoError(t, err)
		assert.NoError(t, tempNodeConf.Truncate(0))
		_, err = tempNodeConf.Write([]byte(tt.inDb))
		assert.NoError(t, err)
		assert.NoError(t, tempNodeConf.Sync())

		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			verifyOutput(t, baseCmd, tt.stdout)
		})

		t.Run(tt.name+" output to yaml format", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-o", "yaml"}
			baseCmd.SetArgs(append(args, tt.args...))
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "nodes:\n- nodeentry:\n    nodesimple:\n      nodename: n01\n      profiles: default\n      network: \"\"\n")
		})

		t.Run(tt.name+" output to json format", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-o", "json"}
			baseCmd.SetArgs(append(args, tt.args...))
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "{\"nodes\":[{\"NodeEntry\":{\"NodeSimple\":{\"node_name\":\"n01\",\"profiles\":\"default\"}}")
		})

		t.Run(tt.name+" output to csv format", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-o", "csv"}
			baseCmd.SetArgs(append(args, tt.args...))
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			verifyOutput(t, baseCmd, "NODENAME,PROFILES,NETWORK\nn01,default,\n")
		})

		t.Run(tt.name+" output to text format", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-o", "text"}
			baseCmd.SetArgs(append(args, tt.args...))
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			verifyOutput(t, baseCmd, "NODENAMEPROFILESNETWORK\nn01default\n")
		})

		// test with other flags, only needing to test the headers
		t.Run(tt.name+" output to csv format (all view)", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-a", "-o", "csv"}
			baseCmd.SetArgs(append(args, tt.args...))
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			verifyOutput(t, baseCmd, "NODE,FIELD,PROFILE,VALUE\n")
		})

		t.Run(tt.name+" output to csv format (full all view)", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-A", "-o", "csv"}
			baseCmd.SetArgs(append(args, tt.args...))
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			verifyOutput(t, baseCmd, "NODE,FIELD,PROFILE,VALUE\n")
		})

		t.Run(tt.name+" output to csv format (ipmi view)", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-i", "-o", "csv"}
			baseCmd.SetArgs(append(args, tt.args...))
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			verifyOutput(t, baseCmd, "NODENAME,IPMIIPADDR,IPMIPORT,IPMIUSERNAME,IPMIINTERFACE,IPMIESCAPECHAR\n")
		})

		t.Run(tt.name+" output to csv format (long view)", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-l", "-o", "csv"}
			baseCmd.SetArgs(append(args, tt.args...))
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			verifyOutput(t, baseCmd, "NODENAME,KERNELOVERRIDE,CONTAINER,OVERLAYS(S/R)\n")
		})

		t.Run(tt.name+" output to csv format (network view)", func(t *testing.T) {
			baseCmd := GetCommand()
			args := []string{"-n", "-o", "csv"}
			baseCmd.SetArgs(append(args, tt.args...))
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			verifyOutput(t, baseCmd, "NODENAME,NAME,HWADDR,IPADDR,GATEWAY,DEVICE\n")
		})
	}
}

func verifyOutput(t *testing.T, baseCmd *cobra.Command, content string) {
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	err := baseCmd.Execute()
	assert.NoError(t, err)

	stdoutC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR)
		stdoutC <- buf.String()
	}()
	stdoutW.Close()

	stdout := <-stdoutC
	assert.NotEmpty(t, stdout, "output should not be empty")
	assert.Contains(t, strings.ReplaceAll(stdout, " ", ""), strings.ReplaceAll(content, " ", ""))
}
