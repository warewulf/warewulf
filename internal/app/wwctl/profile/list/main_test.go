package list

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
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
	}

	conf_yml := `
WW_INTERNAL: 0
    `

	conf := warewulfconf.New()
	err := conf.Read([]byte(conf_yml))
	assert.NoError(t, err)
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		_, err = node.TestNew([]byte(tt.inDb))
		assert.NoError(t, err)
		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			stdoutR, stdoutW, _ := os.Pipe()
			os.Stdout = stdoutW
			err = baseCmd.Execute()
			if err != nil {
				t.Errorf("Received error when running command, err: %v", err)
				t.FailNow()
			}
			stdoutC := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, stdoutR)
				stdoutC <- buf.String()
			}()
			stdoutW.Close()

			stdout := <-stdoutC
			stdout = strings.TrimSpace(stdout)
			stdout = strings.ReplaceAll(stdout, " ", "")
			assert.NotEmpty(t, stdout, "os.stdout should not be empty")
			tt.stdout = strings.ReplaceAll(strings.TrimSpace(tt.stdout), " ", "")
			if stdout != strings.ReplaceAll(strings.TrimSpace(tt.stdout), " ", "") {
				t.Errorf("Got wrong output, got:\n '%s'\n, but want:\n '%s'\n", stdout, tt.stdout)
				t.FailNow()
			}
		})
	}
}
