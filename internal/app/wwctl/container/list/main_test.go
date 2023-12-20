package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_List_Args(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdout   string
		inDb     string
		mockFunc func()
	}{
		{
			name: "container list test",
			args: []string{"-l"},
			stdout: `  CONTAINER NAME  NODES  KERNEL VERSION  CREATION TIME        MODIFICATION TIME    SIZE
 test            1      kernel`,
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

	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		env := testenv.New(t)
		env.WriteFile(t, "etc/warewulf/nodes.conf", tt.inDb)

		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			buf := new(bytes.Buffer)
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			stdoutR, stdoutW, _ := os.Pipe()
			os.Stdout = stdoutW
			wwlog.SetLogWriter(os.Stdout)
			baseCmd.SetOut(os.Stdout)
			baseCmd.SetErr(os.Stdout)
			err := baseCmd.Execute()
			if tt.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			stdoutC := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, stdoutR)
				stdoutC <- buf.String()
			}()
			stdoutW.Close()
			stdout := <-stdoutC
			assert.Equal(t, tt.output, stdout)
			assert.Equal(t,
				strings.ReplaceAll(strings.TrimSpace(tt.output), " ", ""),
				strings.ReplaceAll(strings.TrimSpace(stdout), " ", ""))

		})
	}
}
