package list

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_List(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdout   string
		inDb     string
		mockFunc func()
	}{
		{
			name:   "container list test",
			args:   []string{},
			stdout: `test            1      kernel`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`,
			mockFunc: func() {
				containerList = func() (containerInfo []*wwapiv1.ContainerInfo, err error) {
					containerInfo = append(containerInfo, &wwapiv1.ContainerInfo{
						Name:          "test",
						NodeCount:     1,
						KernelVersion: "kernel",
						CreateDate:    uint64(time.Unix(0, 0).Unix()),
						ModDate:       uint64(time.Unix(0, 0).Unix()),
						Size:          uint64(1),
					})
					return
				}
			},
		},
	}

	conf_yml := `
WW_INTERNAL: 0
    `

	conf := warewulfconf.Get()
	err := conf.Parse([]byte(conf_yml))
	assert.NoError(t, err)
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		_, err = node.Parse([]byte(tt.inDb))
		assert.NoError(t, err)
		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
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
			assert.NotEmpty(t, stdout, "os.stdout should not be empty")
			if !strings.Contains(stdout, tt.stdout) {
				t.Errorf("Got wrong output, got:\n '%s'\n, but want:\n '%s'\n", stdout, tt.stdout)
				t.FailNow()
			}
		})
	}
}
