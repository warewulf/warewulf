package list

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
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
  default       --`,
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
			name: "profile list returns multiple profiles",
			args: []string{"default,test"},
			stdout: `PROFILE NAME  COMMENT/DESCRIPTION
  default       --
  test          --`,
			inDb: `WW_INTERNAL: 45
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
			name: "profile list returns one profiles",
			args: []string{"test,"},
			stdout: `PROFILE NAME  COMMENT/DESCRIPTION
  test          --`,
			inDb: `WW_INTERNAL: 45
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
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
  test: {}
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
		assert.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			verifyOutput(t, baseCmd, tt.stdout)
		})
	}
}

func TestListMultipleFormats(t *testing.T) {
	t.Skip("temporally skip this test")
	tests := []struct {
		name   string
		args   []string
		output []string
		inDb   string
	}{
		{
			name:   "single profile list yaml output",
			args:   []string{"-y"},
			output: []string{"default:\n  AssetKey: \"\"\n  ClusterName: \"\"\n  Comment: \"\"\n  ContainerName: \"\"\n  Discoverable: \"\"\n  Disks: {}\n  FileSystems: {}\n  Grub: \"\"\n  Id: |\n    Source: explicit\n    Value: default\n  Init: \"\"\n  Ipmi:\n    EscapeChar: \"\"\n    Gateway: \"\"\n    Interface: \"\"\n    Ipaddr: \"\"\n    Netmask: \"\"\n    Password: \"\"\n    Port: \"\"\n    Tags: null\n    UserName: \"\"\n    Write: \"\"\n  Ipxe: \"\"\n  Kernel:\n    Args: \"\"\n    Override: \"\"\n  NetDevs: {}\n  PrimaryNetDev: \"\"\n  Profiles: \"\"\n  Root: \"\"\n  RuntimeOverlay: \"\"\n  SystemOverlay: \"\"\n  Tags: {}\n"},
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
			name:   "single profile list json output",
			args:   []string{"-j"},
			output: []string{"{\"default\":{\"Id\":\"Source: explicit\\nValue: default\\n\",\"Comment\":\"\",\"ClusterName\":\"\",\"ContainerName\":\"\",\"Ipxe\":\"\",\"Grub\":\"\",\"RuntimeOverlay\":\"\",\"SystemOverlay\":\"\",\"Root\":\"\",\"Discoverable\":\"\",\"Init\":\"\",\"AssetKey\":\"\",\"Kernel\":{\"Override\":\"\",\"Args\":\"\"},\"Ipmi\":{\"Ipaddr\":\"\",\"Netmask\":\"\",\"Port\":\"\",\"Gateway\":\"\",\"UserName\":\"\",\"Password\":\"\",\"Interface\":\"\",\"EscapeChar\":\"\",\"Write\":\"\",\"Tags\":null},\"Profiles\":\"\",\"PrimaryNetDev\":\"\",\"NetDevs\":{},\"Tags\":{},\"Disks\":{},\"FileSystems\":{}}}\n"},
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
			name:   "multiple profiles list yaml output",
			args:   []string{"-y"},
			output: []string{"default", "test"},
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
			name:   "multiple profiles list json output",
			args:   []string{"-j"},
			output: []string{"default", "test"},
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

			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			for _, output := range tt.output {
				assert.Contains(t, buf.String(), output)
			}
		})
	}
}

func verifyOutput(t *testing.T, baseCmd *cobra.Command, content string) {
	stdoutR, stdoutW, _ := os.Pipe()
	oriout := os.Stdout
	os.Stdout = stdoutW
	wwlog.SetLogWriter(os.Stdout)
	baseCmd.SetOut(os.Stdout)
	baseCmd.SetErr(os.Stdout)
	err := baseCmd.Execute()
	assert.NoError(t, err)

	stdoutC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR)
		stdoutC <- buf.String()
	}()
	stdoutW.Close()
	os.Stdout = oriout

	stdout := <-stdoutC
	stdout = strings.ReplaceAll(strings.TrimSpace(stdout), " ", "")
	assert.NotEmpty(t, stdout, "output should not be empty")
	content = strings.ReplaceAll(strings.TrimSpace(content), " ", "")
	assert.Contains(t, stdout, strings.ReplaceAll(strings.TrimSpace(content), " ", ""))
}
