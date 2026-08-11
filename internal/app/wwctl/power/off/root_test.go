package off

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
	env.ImportFile("usr/share/warewulf/bmc/ipmitool.tmpl", "../../../../../lib/warewulf/bmc/ipmitool.tmpl")

	tests := map[string]struct {
		args     []string
		expected string
	}{
		"power off": {
			args:     []string{"--show", "n01"},
			expected: `10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power off`,
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

// Test_PowerOff_NoIpmiAddress checks that a node with no ipmi ipaddr, such as a
// virtual machine with no BMC, is skipped rather than acted on locally
func Test_PowerOff_NoIpmiAddress(t *testing.T) {
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
      ipaddr: 10.10.10.10
  n02:
    profiles:
    - default
    ipmi: {}`)
	env.ImportFile("usr/share/warewulf/bmc/ipmitool.tmpl", "../../../../../lib/warewulf/bmc/ipmitool.tmpl")

	tests := map[string]struct {
		args     []string
		expected string
	}{
		"only a node without an address": {
			args: []string{"--show", "n02"},
		},
		"a node with an address and one without": {
			args:     []string{"--show", "n01,n02"},
			expected: `10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power off`,
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

			assert.Error(t, err)
			assert.Contains(t, buf.String(), "n02: No IPMI IP address")
			assert.NotContains(t, buf.String(), `ipmitool -U "admin" -P "admin" chassis power off`)
			if tt.expected != "" {
				assert.Contains(t, buf.String(), tt.expected)
			}
		})
	}
}
