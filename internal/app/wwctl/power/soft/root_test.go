package soft

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Power_Status(t *testing.T) {
	daemon.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `
nodeprofiles:
  default:
    ipmi:
      template: ipmitool.tmpl
      username: admin
      password: admin
nodes:
  n01:
    profiles:
    - default
    ipmi:
      ipaddr: 10.10.10.10`)
	env.ImportFile("usr/share/warewulf/bmc/ipmitool.tmpl", "../../../../../lib/warewulf/bmc/ipmitool.tmpl")

	tests := map[string]struct {
		args     []string
		expected string
	}{
		"power soft": {
			args:     []string{"--show", "n01"},
			expected: "10.10.10.10: ipmitool -I lan -H 10.10.10.10 -p 623 -U admin -P admin -e ~ chassis power soft",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			baseCmd := GetCommand()
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			baseCmd.SetArgs(tt.args)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.expected), strings.TrimSpace(buf.String()))
		})
	}
}
