package poweron

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Power_Status(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll(t)

	env.WriteFile(t, "etc/warewulf/nodes.conf", `WW_INTERNAL: 43
nodeprofiles:
  default:
    ipmi:
      username: admin
      password: admin
nodes:
  n01:
    profiles:
    - default
    ipmi:
      ipaddr: 10.10.10.10`)
	warewulfd.SetNoDaemon()
	t.Run("ipmitool status test", func(t *testing.T) {
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		baseCmd.SetArgs([]string{"--show", "n01"})
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Equal(t, "10.10.10.10: ipmitool -I lan -H 10.10.10.10 -p 623 -U admin -P admin -e ~ chassis power on", strings.TrimSpace(buf.String()))
	})
}
