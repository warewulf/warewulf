package reset

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
		"power reset": {
			args:     []string{"--show", "n01"},
			expected: `10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power reset`,
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

func Test_Power_Reset_Groups(t *testing.T) {
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
    nodegroups:
    - admin
    ipmi:
      ipaddr: 10.10.10.11
  n03:
    profiles:
    - default
    ipmi:
      ipaddr: 10.10.10.12
nodegroups:
  rack1:
    - n01
    - n02`)
	env.ImportFile("usr/share/warewulf/bmc/ipmitool.tmpl", "../../../../../lib/warewulf/bmc/ipmitool.tmpl")

	tests := map[string]struct {
		args     []string
		expected []string
	}{
		"@rack1 expands to nodegroups stanza": {
			args: []string{"--show", "@rack1"},
			expected: []string{
				`10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power reset`,
				`10.10.10.11: ipmitool -H 10.10.10.11 -U "admin" -P "admin" chassis power reset`,
			},
		},
		"@admin expands to per-node nodegroups field": {
			args: []string{"--show", "@admin"},
			expected: []string{
				`10.10.10.11: ipmitool -H 10.10.10.11 -U "admin" -P "admin" chassis power reset`,
			},
		},
		"@all expands to every node": {
			args: []string{"--show", "@all"},
			expected: []string{
				`10.10.10.10: ipmitool -H 10.10.10.10 -U "admin" -P "admin" chassis power reset`,
				`10.10.10.11: ipmitool -H 10.10.10.11 -U "admin" -P "admin" chassis power reset`,
				`10.10.10.12: ipmitool -H 10.10.10.12 -U "admin" -P "admin" chassis power reset`,
			},
		},
		"mix plain and group dedupes": {
			args: []string{"--show", "n01", "@admin"},
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
