package list

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
			inDb: `WW_INTERNAL: 43
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
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			baseCmd := GetCommand()
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			baseCmd.SetArgs(tt.args)
			verifyOutput(t, baseCmd, tt.stdout)
		})

		t.Run(tt.name+" with output yaml", func(t *testing.T) {
			tt.mockFunc()
			baseCmd := GetCommand()
			args := tt.args
			baseCmd.SetArgs(append(args, "-o", "yaml"))
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "Containers:\n  test:\n  - Nodes: 1\n    KernelVersion: kernel\n    CreationTime: 0\n    ModificationTime: 0\n    Size: 1\n")
		})

		t.Run(tt.name+" with output json", func(t *testing.T) {
			tt.mockFunc()
			baseCmd := GetCommand()
			args := tt.args
			baseCmd.SetArgs(append(args, "-o", "json"))
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "{\"Containers\":{\"test\":[{\"Nodes\":1,\"KernelVersion\":\"kernel\",\"CreationTime\":0,\"ModificationTime\":0,\"Size\":1}]}}\n")
		})

		t.Run(tt.name+" with output csv", func(t *testing.T) {
			tt.mockFunc()
			baseCmd := GetCommand()
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			args := tt.args
			baseCmd.SetArgs(append(args, "-o", "csv"))
			assert.NoError(t, err)
			verifyOutput(t, baseCmd, "CONTAINER NAME,NODES,KERNEL VERSION,CREATION TIME,MODIFICATION TIME,SIZE\ntest,1,kernel,01 Jan 70 00:00 UTC,01 Jan 70 00:00 UTC,1 B\n")
		})

		t.Run(tt.name+" with output text", func(t *testing.T) {
			tt.mockFunc()
			baseCmd := GetCommand()
			baseCmd.SetOut(nil)
			baseCmd.SetErr(nil)
			args := tt.args
			baseCmd.SetArgs(append(args, "-o", "text"))
			assert.NoError(t, err)
			verifyOutput(t, baseCmd, "  CONTAINER NAME  NODES  KERNEL VERSION  CREATION TIME        MODIFICATION TIME    SIZE  \n  test            1      kernel          01 Jan 70 00:00 UTC  01 Jan 70 00:00 UTC  1 B   \n")
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
	assert.Contains(t, stdout, content)
}
