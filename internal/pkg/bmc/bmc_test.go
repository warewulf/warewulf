package bmc

import (
	"net"
	"strings"
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
		"ipmitool PowerStatus empty": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template: "ipmitool.tmpl",
				},
			},
			cmdStr: `ipmitool chassis power status`,
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

func Test_VirtualBMC_Kind(t *testing.T) {
	tests := map[string]struct {
		bmc         TemplateStruct
		err         bool
		containsStr []string
	}{
		"kind PowerOn default": {
			bmc: TemplateStruct{
				Cmd: "PowerOn",
				IpmiConf: node.IpmiConf{
					Template: "kind.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"docker inspect 192-168-1-100",
				"docker start 192-168-1-100",
				"docker run -d",
				"--name 192-168-1-100",
				"--cpus 2",
				"--memory 2g",
				"kindest/node:latest",
			},
		},
		"kind PowerOn custom": {
			bmc: TemplateStruct{
				Cmd: "PowerOn",
				IpmiConf: node.IpmiConf{
					Template: "kind.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
					Tags: map[string]string{
						"nodename": "testnode",
						"cpu":      "4",
						"memory":   "8g",
						"disk":     "50g",
						"image":    "ubuntu:22.04",
					},
				},
			},
			containsStr: []string{
				"docker inspect testnode",
				"docker start testnode",
				"--name testnode",
				"--cpus 4",
				"--memory 8g",
				"ubuntu:22.04",
			},
		},
		"kind PowerOff": {
			bmc: TemplateStruct{
				Cmd: "PowerOff",
				IpmiConf: node.IpmiConf{
					Template: "kind.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"docker inspect 192-168-1-100",
				"docker stop 192-168-1-100",
			},
		},
		"kind PowerStatus": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template: "kind.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"docker inspect 192-168-1-100",
				"echo \"ON\"",
				"echo \"OFF\"",
				"echo \"NOT_EXIST\"",
			},
		},
		"kind PowerCycle": {
			bmc: TemplateStruct{
				Cmd: "PowerCycle",
				IpmiConf: node.IpmiConf{
					Template: "kind.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"docker inspect 192-168-1-100",
				"docker restart 192-168-1-100",
			},
		},
		"kind SDRList": {
			bmc: TemplateStruct{
				Cmd: "SDRList",
				IpmiConf: node.IpmiConf{
					Template: "kind.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
					Tags: map[string]string{
						"cpu":    "4",
						"memory": "8g",
						"disk":   "50g",
					},
				},
			},
			containsStr: []string{
				"echo \"CPU: 4\"",
				"echo \"Memory: 8g\"",
				"echo \"Disk: 50g\"",
			},
		},
	}

	for name, test := range tests {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.ImportFile("usr/share/warewulf/bmc/kind.tmpl", "../../../lib/warewulf/bmc/kind.tmpl")

		t.Run(name, func(t *testing.T) {
			cmdStr, err := test.bmc.getCommand()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, substr := range test.containsStr {
					assert.Contains(t, cmdStr, substr, "Command should contain: %s", substr)
				}
			}
		})
	}
}

func Test_VirtualBMC_KindLibvirt(t *testing.T) {
	tests := map[string]struct {
		bmc         TemplateStruct
		err         bool
		containsStr []string
	}{
		"libvirt PowerOn default": {
			bmc: TemplateStruct{
				Cmd: "PowerOn",
				IpmiConf: node.IpmiConf{
					Template: "kind-libvirt.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"virsh dominfo 192-168-1-100",
				"virsh start 192-168-1-100",
				"qemu-img create -f qcow2",
				"virt-install",
				"--name 192-168-1-100",
				"--vcpus 2",
				"--memory 2048",
			},
		},
		"libvirt PowerOn custom": {
			bmc: TemplateStruct{
				Cmd: "PowerOn",
				IpmiConf: node.IpmiConf{
					Template: "kind-libvirt.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
					Tags: map[string]string{
						"nodename":   "testvm",
						"cpu":        "8",
						"memory":     "16384",
						"disk":       "100",
						"disk_path":  "/custom/path",
						"os_variant": "ubuntu22.04",
					},
				},
			},
			containsStr: []string{
				"virsh dominfo testvm",
				"--name testvm",
				"--vcpus 8",
				"--memory 16384",
				"/custom/path/testvm.qcow2",
				"--os-variant ubuntu22.04",
			},
		},
		"libvirt PowerOff": {
			bmc: TemplateStruct{
				Cmd: "PowerOff",
				IpmiConf: node.IpmiConf{
					Template: "kind-libvirt.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"virsh dominfo 192-168-1-100",
				"virsh shutdown 192-168-1-100",
			},
		},
		"libvirt PowerStatus": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template: "kind-libvirt.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"virsh dominfo 192-168-1-100",
				"virsh domstate 192-168-1-100",
				"echo \"ON\"",
				"echo \"OFF\"",
			},
		},
		"libvirt SDRList": {
			bmc: TemplateStruct{
				Cmd: "SDRList",
				IpmiConf: node.IpmiConf{
					Template: "kind-libvirt.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"virsh dominfo 192-168-1-100",
				"qemu-img info",
			},
		},
	}

	for name, test := range tests {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.ImportFile("usr/share/warewulf/bmc/kind-libvirt.tmpl", "../../../lib/warewulf/bmc/kind-libvirt.tmpl")

		t.Run(name, func(t *testing.T) {
			cmdStr, err := test.bmc.getCommand()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, substr := range test.containsStr {
					assert.Contains(t, cmdStr, substr, "Command should contain: %s", substr)
				}
			}
		})
	}
}

func Test_VirtualBMC_KindQemu(t *testing.T) {
	tests := map[string]struct {
		bmc         TemplateStruct
		err         bool
		containsStr []string
	}{
		"qemu PowerOn default": {
			bmc: TemplateStruct{
				Cmd: "PowerOn",
				IpmiConf: node.IpmiConf{
					Template: "kind-qemu.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"/var/run/qemu-192-168-1-100.pid",
				"qemu-img create -f qcow2",
				"qemu-system-x86_64",
				"-name 192-168-1-100",
				"-m 2048",
				"-smp 2",
				"/var/lib/qemu/images/192-168-1-100.qcow2",
			},
		},
		"qemu PowerOn custom": {
			bmc: TemplateStruct{
				Cmd: "PowerOn",
				IpmiConf: node.IpmiConf{
					Template: "kind-qemu.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
					Tags: map[string]string{
						"nodename":  "qemutest",
						"cpu":       "4",
						"memory":    "8192",
						"disk":      "80",
						"disk_path": "/var/qemu",
						"mac":       "52:54:00:12:34:56",
					},
				},
			},
			containsStr: []string{
				"/var/run/qemu-qemutest.pid",
				"-name qemutest",
				"-m 8192",
				"-smp 4",
				"/var/qemu/qemutest.qcow2",
				"mac=52:54:00:12:34:56",
			},
		},
		"qemu PowerOff": {
			bmc: TemplateStruct{
				Cmd: "PowerOff",
				IpmiConf: node.IpmiConf{
					Template: "kind-qemu.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"/var/run/qemu-192-168-1-100.pid",
				"system_powerdown",
				"/var/run/qemu-192-168-1-100.monitor",
			},
		},
		"qemu PowerStatus": {
			bmc: TemplateStruct{
				Cmd: "PowerStatus",
				IpmiConf: node.IpmiConf{
					Template: "kind-qemu.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"/var/run/qemu-192-168-1-100.pid",
				"echo \"ON\"",
				"echo \"OFF\"",
			},
		},
		"qemu PowerReset": {
			bmc: TemplateStruct{
				Cmd: "PowerReset",
				IpmiConf: node.IpmiConf{
					Template: "kind-qemu.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
				},
			},
			containsStr: []string{
				"system_reset",
				"/var/run/qemu-192-168-1-100.monitor",
			},
		},
		"qemu SDRList": {
			bmc: TemplateStruct{
				Cmd: "SDRList",
				IpmiConf: node.IpmiConf{
					Template: "kind-qemu.tmpl",
					Ipaddr:   net.IP{192, 168, 1, 100},
					Tags: map[string]string{
						"cpu":    "8",
						"memory": "16384",
					},
				},
			},
			containsStr: []string{
				"echo \"CPU: 8\"",
				"echo \"Memory: 16384MB\"",
				"qemu-img info",
			},
		},
	}

	for name, test := range tests {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.ImportFile("usr/share/warewulf/bmc/kind-qemu.tmpl", "../../../lib/warewulf/bmc/kind-qemu.tmpl")

		t.Run(name, func(t *testing.T) {
			cmdStr, err := test.bmc.getCommand()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, substr := range test.containsStr {
					assert.Contains(t, cmdStr, substr, "Command should contain: %s", substr)
				}
			}
		})
	}
}

func Test_AllVirtualBMC_AllCommands(t *testing.T) {
	templates := []string{"kind.tmpl", "kind-libvirt.tmpl", "kind-qemu.tmpl"}
	commands := []string{"PowerOn", "PowerOff", "PowerCycle", "PowerReset", "PowerSoft", "PowerStatus", "SDRList", "SensorList"}

	env := testenv.New(t)
	defer env.RemoveAll()
	env.ImportFile("usr/share/warewulf/bmc/kind.tmpl", "../../../lib/warewulf/bmc/kind.tmpl")
	env.ImportFile("usr/share/warewulf/bmc/kind-libvirt.tmpl", "../../../lib/warewulf/bmc/kind-libvirt.tmpl")
	env.ImportFile("usr/share/warewulf/bmc/kind-qemu.tmpl", "../../../lib/warewulf/bmc/kind-qemu.tmpl")

	for _, tmpl := range templates {
		for _, cmd := range commands {
			testName := tmpl + "_" + cmd
			t.Run(testName, func(t *testing.T) {
				bmc := TemplateStruct{
					Cmd: cmd,
					IpmiConf: node.IpmiConf{
						Template: tmpl,
						Ipaddr:   net.IP{10, 0, 0, 1},
						Tags: map[string]string{
							"nodename": "testnode",
						},
					},
				}

				cmdStr, err := bmc.getCommand()
				assert.NoError(t, err, "Template %s with command %s should not error", tmpl, cmd)
				assert.NotEmpty(t, cmdStr, "Template %s with command %s should produce non-empty command", tmpl, cmd)

				// Verify the command contains the node name
				assert.True(t,
					strings.Contains(cmdStr, "testnode") || strings.Contains(cmdStr, "10-0-0-1"),
					"Command should reference the node: %s", cmdStr)
			})
		}
	}
}
