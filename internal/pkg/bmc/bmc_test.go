package bmc

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_Ipmitool(t *testing.T) {
	tests := map[string]struct {
		bmc    TemplateStruct
		err    bool
		cmdStr string
	}{
		"no template": {
			bmc: TemplateStruct{},
			err: true,
		},
		// with no ipaddr the command has no -H and would target the local bmc
		"ipmitool PowerStatus empty": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template: "ipmitool.tmpl",
				},
			},
			err: true,
		},
		// IsUnspecified is true for 0.0.0.0 but false for nil, so cover both
		"ipmitool PowerStatus unspecified ipaddr": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template: "ipmitool.tmpl",
					Ipaddr:   net.IPv4zero,
					UserName: "root",
					Password: "calvin",
				},
			},
			cmdStr: `ipmitool -H 0.0.0.0 -U "root" -P "calvin" chassis power status`,
		},
		// an explicit interface is an intentional choice of transport, so it is
		// honoured even without an address
		"ipmitool PowerStatus explicit interface": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template:  "ipmitool.tmpl",
					Interface: "open",
				},
			},
			cmdStr: `ipmitool -I open chassis power status`,
		},
		"ipmitool PowerStatus full": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template:   "ipmitool.tmpl",
					Interface:  "lanplus",
					EscapeChar: "~",
					Port:       "687",
					Ipaddr:     net.IP{192, 168, 1, 100},
					UserName:   "root",
					Password:   "calvin",
				},
			},
			cmdStr: `ipmitool -I lanplus -e "~" -p 687 -H 192.168.1.100 -U "root" -P "calvin" chassis power status`,
		},
		"nobmc PowerStatus full": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template:   "nobmc.tmpl",
					Interface:  "lanplus",
					EscapeChar: "~",
					Port:       "687",
					Ipaddr:     net.IP{192, 168, 1, 100},
					UserName:   "root",
					Password:   "calvin",
				},
			},
			cmdStr: `ping -c 1 "192.168.1.100" &> /dev/null && echo ON || echo OFF`,
		},
	}

	for name, test := range tests {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.ImportFile("usr/share/warewulf/bmc/ipmitool.tmpl", "../../../lib/warewulf/bmc/ipmitool.tmpl")
		env.ImportFile("usr/share/warewulf/bmc/nobmc.tmpl", "../../../lib/warewulf/bmc/nobmc.tmpl")

		t.Run(name, func(t *testing.T) {
			cmdStr, err := test.bmc.getCommand()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if test.cmdStr != "" {
				assert.Equal(t, test.cmdStr, cmdStr)
			}
		})
	}
}
