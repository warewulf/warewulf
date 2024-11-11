package list

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

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
			stdout: `
PROFILE NAME  COMMENT/DESCRIPTION
------------  -------------------
default       --
`,
			inDb: `
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
			stdout: `
PROFILE NAME  COMMENT/DESCRIPTION
------------  -------------------
default       --
test          --
`,
			inDb: `
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
			name: "profile list returns one profile",
			args: []string{"test,"},
			stdout: `
PROFILE NAME  COMMENT/DESCRIPTION
------------  -------------------
test          --
`,
			inDb: `
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
			stdout: `
PROFILE NAME  COMMENT/DESCRIPTION
------------  -------------------
default       --
test          --
`,
			inDb: `
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

	conf_yml := ``
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
			assert.Equal(t, strings.TrimSpace(tt.stdout), strings.TrimSpace(stdout))
		})
	}
}

func TestListMultipleFormats(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		output string
		inDb   string
	}{
		{
			name:   "single profile list yaml output",
			args:   []string{"-y"},
			output: `default: {}`,
			inDb: `
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`,
		},
		{
			name: "single profile list json output",
			args: []string{"-j"},
			output: `
{
  "default": {
    "Comment": "",
    "ClusterName": "",
    "ContainerName": "",
    "Ipxe": "",
    "RuntimeOverlay": null,
    "SystemOverlay": null,
    "Kernel": null,
    "Ipmi": null,
    "Init": "",
    "Root": "",
    "NetDevs": null,
    "Tags": null,
    "PrimaryNetDev": "",
    "Disks": null,
    "FileSystems": null
  }
}
`,
			inDb: `
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`,
		},
		{
			name: "multiple profiles list yaml output",
			args: []string{"-y"},
			output: `
default: {}
test: {}
`,
			inDb: `
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
			name: "multiple profiles list json output",
			args: []string{"-j"},
			output: `
{
  "default": {
    "Comment": "",
    "ClusterName": "",
    "ContainerName": "",
    "Ipxe": "",
    "RuntimeOverlay": null,
    "SystemOverlay": null,
    "Kernel": null,
    "Ipmi": null,
    "Init": "",
    "Root": "",
    "NetDevs": null,
    "Tags": null,
    "PrimaryNetDev": "",
    "Disks": null,
    "FileSystems": null
  },
  "test": {
    "Comment": "",
    "ClusterName": "",
    "ContainerName": "",
    "Ipxe": "",
    "RuntimeOverlay": null,
    "SystemOverlay": null,
    "Kernel": null,
    "Ipmi": null,
    "Init": "",
    "Root": "",
    "NetDevs": null,
    "Tags": null,
    "PrimaryNetDev": "",
    "Disks": null,
    "FileSystems": null
  }
}
`,
			inDb: `
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

	conf_yml := ``
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
			assert.Equal(t, strings.TrimSpace(tt.output), strings.TrimSpace(buf.String()))
		})
	}
}
