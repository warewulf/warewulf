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
	}
}

func Test_List_Multiple_Format(t *testing.T) {
	const (
		YAML = iota
		JSON
		CSV
		TEXT
	)

	tests := []struct {
		name       string
		args       []string
		outputType int
		output     string
		inDb       string
	}{
		{
			name:       "single node list yaml output",
			args:       []string{"-o", "yaml"},
			outputType: YAML,
			output:     "Nodes:\n  n01:\n  - Profiles: default\n    Network: \"\"\n",
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
			name:       "single node list json output",
			args:       []string{"-o", "json"},
			outputType: JSON,
			output:     "{\"Nodes\":{\"n01\":[{\"Profiles\":\"default\",\"Network\":\"\"}]}}\n",
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
			name:       "single node list csv output",
			args:       []string{"-o", "csv"},
			outputType: CSV,
			output:     "NODENAME,PROFILES,NETWORK\nn01,default,\n",
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
			name:       "single node list text output",
			args:       []string{"-o", "text"},
			outputType: TEXT,
			output:     "NODENAMEPROFILESNETWORK\nn01default\n",
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
			name:       "single node list csv output (all view)",
			args:       []string{"-a", "-o", "csv"},
			outputType: CSV,
			output:     "NODE,FIELD,PROFILE,VALUE\nn01,Id,--,n01\nn01,Ipxe,--,(default)\nn01,RuntimeOverlay,--,(generic)\nn01,SystemOverlay,--,(wwinit)\nn01,Root,--,(initramfs)\nn01,Init,--,(/sbin/init)\nn01,KernelArgs,--,(quietcrashkernel=novga=791net.naming-scheme=v238)\nn01,Profiles,--,default\n",
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
			name:       "single node list csv output (full all view)",
			args:       []string{"-A", "-o", "csv"},
			outputType: CSV,
			output:     "NODE,FIELD,PROFILE,VALUE\nn01,Id,--,n01\nn01,Comment,--,--\nn01,ClusterName,--,--\nn01,ContainerName,--,--\nn01,Ipxe,--,(default)\nn01,Grub,--,--\nn01,RuntimeOverlay,--,(generic)\nn01,SystemOverlay,--,(wwinit)\nn01,Root,--,(initramfs)\nn01,Discoverable,--,--\nn01,Init,--,(/sbin/init)\nn01,AssetKey,--,--\nn01,KernelOverride,--,--\nn01,KernelArgs,--,(quietcrashkernel=novga=791net.naming-scheme=v238)\nn01,IpmiIpaddr,--,--\nn01,IpmiNetmask,--,--\nn01,IpmiPort,--,--\nn01,IpmiGateway,--,--\nn01,IpmiUserName,--,--\nn01,IpmiPassword,--,--\nn01,IpmiInterface,--,--\nn01,IpmiEscapeChar,--,--\nn01,IpmiWrite,--,--\nn01,IpmiTags[],,\nn01,Profiles,--,default\nn01,PrimaryNetDev,--,--\nn01,NetDevs[],,\nn01,Tags[],,\nn01,Disks[],,\nn01,FileSystems[],,\n",
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
			name:       "single node list csv output (ipmi view)",
			args:       []string{"-i", "-o", "csv"},
			outputType: CSV,
			output:     "NODENAME,IPMIIPADDR,IPMIPORT,IPMIUSERNAME,IPMIINTERFACE,IPMIESCAPECHAR\nn01,--,--,--,--,--\n",
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
			name:       "single node list csv output (long view)",
			args:       []string{"-l", "-o", "csv"},
			outputType: CSV,
			output:     "NODENAME,KERNELOVERRIDE,CONTAINER,OVERLAYS(S/R)\nn01,--,--,(wwinit)/(generic)\n",
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
			name:       "single node list csv output (network view)",
			args:       []string{"-n", "-o", "csv"},
			outputType: CSV,
			output:     "NODENAME,NAME,HWADDR,IPADDR,GATEWAY,DEVICE\nn01,--,--,--,--,--\n",
			inDb: `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  n01:
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

			if tt.outputType == YAML || tt.outputType == JSON {
				buf := new(bytes.Buffer)
				baseCmd.SetOut(buf)
				baseCmd.SetErr(buf)
				wwlog.SetLogWriter(buf)
				err := baseCmd.Execute()
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tt.output)
			} else {
				baseCmd.SetOut(nil)
				baseCmd.SetErr(nil)
				verifyOutput(t, baseCmd, tt.output)
			}
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
