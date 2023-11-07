package list

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
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
			name: "container list test",
			args: []string{},
			stdout: `
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
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			assert.NoError(t, baseCmd.Execute())
			assert.Contains(t, strings.TrimSpace(buf.String()), strings.TrimSpace(tt.stdout))
		})
	}
}
