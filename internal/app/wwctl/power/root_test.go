package power

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	nodepower "github.com/warewulf/warewulf/internal/app/wwctl/node/power"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Power_Legacy(t *testing.T) {
	warewulfd.SetNoDaemon()
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
	env.ImportFile("usr/share/warewulf/bmc/ipmitool.tmpl", "../../../../lib/warewulf/bmc/ipmitool.tmpl")

	for _, action := range nodepower.Actions {
		t.Run(action.Name, func(t *testing.T) {
			baseCmd := GetCommand()
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			baseCmd.SetArgs([]string{action.Name, "--show", "n01"})
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), `10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power `+action.Name)
			assert.Contains(t, buf.String(), "wwctl node power")
		})
	}
}
