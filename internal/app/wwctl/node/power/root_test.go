package power

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Power(t *testing.T) {
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

	for _, action := range Actions {
		t.Run(action.Name, func(t *testing.T) {
			baseCmd := GetCommand()
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			baseCmd.SetArgs([]string{"--show", action.Name, "n01"})
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Equal(t, `10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power `+action.Name, strings.TrimSpace(buf.String()))
		})
	}
}

func Test_Power_InvalidArgs(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()

	tests := map[string]struct {
		args     []string
		expected string
	}{
		"invalid action": {
			args:     []string{"--show", "bogus", "n01"},
			expected: `invalid action "bogus"`,
		},
		"missing pattern": {
			args:     []string{"--show", "on"},
			expected: "requires at least 2 arg(s)",
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
			assert.Contains(t, err.Error(), tt.expected)
			assert.NotContains(t, buf.String(), "ipmitool")
		})
	}
}

func Test_Power_Groups(t *testing.T) {
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
  rack1:
    groups:
    - rack1
nodes:
  n01:
    profiles:
    - default
    - rack1
    ipmi:
      ipaddr: 10.10.10.10
  n02:
    profiles:
    - default
    - rack1
    groups:
    - admin
    ipmi:
      ipaddr: 10.10.10.11
  n03:
    profiles:
    - default
    ipmi:
      ipaddr: 10.10.10.12`)
	env.ImportFile("usr/share/warewulf/bmc/ipmitool.tmpl", "../../../../../lib/warewulf/bmc/ipmitool.tmpl")

	tests := map[string]struct {
		args     []string
		expected []string
	}{
		"@rack1 expands to profile-inherited group": {
			args: []string{"--show", "reset", "@rack1"},
			expected: []string{
				`10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power reset`,
				`10.10.10.11: ipmitool -H 10.10.10.11 -U "admin" -P "admin" chassis power reset`,
			},
		},
		"@admin expands to per-node groups field": {
			args: []string{"--show", "reset", "@admin"},
			expected: []string{
				`10.10.10.11: ipmitool -H 10.10.10.11 -U "admin" -P "admin" chassis power reset`,
			},
		},
		"@all expands to every node": {
			args: []string{"--show", "reset", "@all"},
			expected: []string{
				`10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power reset`,
				`10.10.10.11: ipmitool -H 10.10.10.11 -U "admin" -P "admin" chassis power reset`,
				`10.10.10.12: ipmitool -H 10.10.10.12 -U "admin" -P "admin" chassis power reset`,
			},
		},
		"mix plain and group dedupes": {
			args: []string{"--show", "reset", "n01", "@admin"},
			expected: []string{
				`10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power reset`,
				`10.10.10.11: ipmitool -H 10.10.10.11 -U "admin" -P "admin" chassis power reset`,
			},
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
			for _, want := range tt.expected {
				assert.Contains(t, buf.String(), want)
			}
		})
	}
}

// Test_Power_NoIpmiAddress checks that a node with no ipmi ipaddr, such as a
// virtual machine with no BMC, is skipped rather than acted on locally
func Test_Power_NoIpmiAddress(t *testing.T) {
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

	for _, action := range Actions {
		tests := map[string]struct {
			args     []string
			expected string
		}{
			"only a node without an address": {
				args: []string{"--show", action.Name, "n02"},
			},
			"a node with an address and one without": {
				args:     []string{"--show", action.Name, "n01,n02"},
				expected: `10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power ` + action.Name,
			},
		}

		for name, tt := range tests {
			t.Run(action.Name+"/"+name, func(t *testing.T) {
				baseCmd := GetCommand()
				buf := new(bytes.Buffer)
				baseCmd.SetOut(buf)
				baseCmd.SetErr(buf)
				wwlog.SetLogWriter(buf)
				baseCmd.SetArgs(tt.args)
				err := baseCmd.Execute()

				assert.Error(t, err)
				assert.Contains(t, buf.String(), "n02: No IPMI IP address")
				assert.NotContains(t, buf.String(), `ipmitool -U "admin" -P "admin" chassis power `+action.Name)
				if tt.expected != "" {
					assert.Contains(t, buf.String(), tt.expected)
				}
			})
		}
	}
}
