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
			name: "profile list returns multiple profiles",
			args: []string{"default,test"},
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
		},
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
			name:       "single profile list yaml output",
			args:       []string{"-o", "yaml"},
			outputType: YAML,
			output:     "Profiles:\n  default:\n  - Comment/Description: --\n",
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
			name:       "single profile list json output",
			args:       []string{"-o", "json"},
			outputType: JSON,
			output:     "{\"Profiles\":{\"default\":[{\"Comment/Description\":\"--\"}]}}\n",
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
			name:       "single profile list csv output",
			args:       []string{"-o", "csv"},
			outputType: CSV,
			output:     "ROFILENAME,COMMENT/DESCRIPTION\ndefault,--",
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
			name:       "single profile list text output",
			args:       []string{"-o", "text"},
			outputType: TEXT,
			output:     "PROFILENAMECOMMENT/DESCRIPTION\ndefault--",
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
			name:       "single profile list yaml output (full view)",
			args:       []string{"-a", "-o", "yaml"},
			outputType: YAML,
			output:     "Profiles:\n  default:\n  - Field: Id\n    Profile: --\n    Value: default\n",
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
			name:       "single profile list json output (full view)",
			args:       []string{"-a", "-o", "json"},
			outputType: JSON,
			output:     "{\"Profiles\":{\"default\":[{\"Field\":\"Id\",\"Profile\":\"--\",\"Value\":\"default\"}]}}\n",
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
			name:       "single profile list csv output (full view)",
			args:       []string{"-a", "-o", "csv"},
			outputType: CSV,
			output:     "PROFILE,FIELD,PROFILE,VALUE\ndefault,Id,--,default",
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
			name:       "single profile list text output (full view)",
			args:       []string{"-a", "-o", "text"},
			outputType: TEXT,
			output:     "PROFILEFIELDPROFILEVALUE\ndefaultId--default",
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
	baseCmd.SetOut(nil)
	baseCmd.SetErr(nil)
	stdoutR, stdoutW, _ := os.Pipe()
	oriout := os.Stdout
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
	os.Stdout = oriout

	stdout := <-stdoutC
	stdout = strings.ReplaceAll(strings.TrimSpace(stdout), " ", "")
	assert.NotEmpty(t, stdout, "output should not be empty")
	content = strings.ReplaceAll(strings.TrimSpace(content), " ", "")
	assert.Contains(t, stdout, strings.ReplaceAll(strings.TrimSpace(content), " ", ""))
}
