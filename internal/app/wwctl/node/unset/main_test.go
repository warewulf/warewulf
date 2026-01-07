package unset

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Node_Unset(t *testing.T) {
	tests := map[string]struct {
		args    []string
		wantErr bool
		inDB    string
		outDB   string
	}{
		// Basic field unsetting
		"unset comment": {
			args:    []string{"--comment", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test comment`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset cluster": {
			args:    []string{"--cluster", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    cluster name: mycluster`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset image": {
			args:    []string{"--image", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    image name: rockylinux-9`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// IP field unsetting - THE MAIN USE CASE!
		"unset netmask (the problem this solves!)": {
			args:    []string{"--netmask", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        netmask: 255.255.255.0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipaddr": {
			args:    []string{"--ipaddr", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset gateway": {
			args:    []string{"--gateway", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        gateway: 192.168.1.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipaddr6": {
			args:    []string{"--ipaddr6", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr6: 2001:db8::1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset gateway6": {
			args:    []string{"--gateway6", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        gateway6: 2001:db8::ffff`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// Multiple IP fields at once
		"unset netmask and ipaddr together": {
			args:    []string{"--netmask", "--ipaddr", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
        netmask: 255.255.255.0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset all IPv4 network fields": {
			args:    []string{"--ipaddr", "--netmask", "--gateway", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
        netmask: 255.255.255.0
        gateway: 192.168.1.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// Network device specific unsetting
		"unset ipaddr on specific network": {
			args:    []string{"--netname", "eth0", "--ipaddr", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
      eth0:
        ipaddr: 10.0.0.100`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100`,
		},
		"unset hwaddr": {
			args:    []string{"--hwaddr", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        hwaddr: "aa:bb:cc:dd:ee:ff"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset netdev": {
			args:    []string{"--netdev", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        device: eth0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset mtu": {
			args:    []string{"--mtu", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        mtu: "9000"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset type": {
			args:    []string{"--type", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        type: ethernet`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// IPMI field unsetting
		"unset ipmiaddr": {
			args:    []string{"--ipmiaddr", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      ipaddr: 192.168.2.100`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipminetmask": {
			args:    []string{"--ipminetmask", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      netmask: 255.255.255.0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipmigateway": {
			args:    []string{"--ipmigateway", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      gateway: 192.168.2.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset all IPMI IP fields": {
			args:    []string{"--ipmiaddr", "--ipminetmask", "--ipmigateway", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      ipaddr: 192.168.2.100
      netmask: 255.255.255.0
      gateway: 192.168.2.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipmiuser": {
			args:    []string{"--ipmiuser", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      username: admin`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipmipass": {
			args:    []string{"--ipmipass", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      password: secret`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipmiport": {
			args:    []string{"--ipmiport", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      port: "623"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipmiinterface": {
			args:    []string{"--ipmiinterface", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      interface: lanplus`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// Kernel fields
		"unset kernelversion": {
			args:    []string{"--kernelversion", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    kernel:
      version: 5.14.0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset kernelargs": {
			args:    []string{"--kernelargs", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    kernel:
      args:
      - quiet
      - splash`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// Overlay fields
		"unset runtime-overlays": {
			args:    []string{"--runtime-overlays", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    runtime overlay:
    - runtime1
    - runtime2`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset system-overlays": {
			args:    []string{"--system-overlays", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    system overlay:
    - system1
    - system2`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// Other fields
		"unset init": {
			args:    []string{"--init", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    init: /sbin/init`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset root": {
			args:    []string{"--root", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    root: /dev/sda1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset ipxe": {
			args:    []string{"--ipxe", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipxe template: custom.ipxe`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset primarynet": {
			args:    []string{"--primarynet", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    primary network: eth0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset profile": {
			args:    []string{"--profile", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: default profile
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default:
    comment: default profile
nodes:
  n01: {}`,
		},
		"unset asset": {
			args:    []string{"--asset", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    asset key: ASSET12345`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// Multiple fields from different categories
		"unset comment, image, and ipaddr": {
			args:    []string{"--comment", "--image", "--ipaddr", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test
    image name: rocky-9
    network devices:
      default:
        ipaddr: 192.168.1.100`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset mixed fields": {
			args:    []string{"--comment", "--ipmiaddr", "--kernelversion", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test node
    ipmi:
      ipaddr: 10.0.0.100
    kernel:
      version: 5.14.0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},

		// Multiple nodes
		"unset on multiple nodes": {
			args:    []string{"--comment", "n0[1-2]"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: node 1
  n02:
    comment: node 2`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}
  n02: {}`,
		},
		"unset netmask on multiple nodes": {
			args:    []string{"--netmask", "n0[1-3]"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        netmask: 255.255.255.0
  n02:
    network devices:
      default:
        netmask: 255.255.255.0
  n03:
    network devices:
      default:
        netmask: 255.255.255.0`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}
  n02: {}
  n03: {}`,
		},

		// Partial unsetting (other fields remain)
		"unset comment but keep other fields": {
			args:    []string{"--comment", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test comment
    image name: rocky-9
    cluster name: mycluster`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    image name: rocky-9
    cluster name: mycluster`,
		},
		"unset netmask but keep ipaddr": {
			args:    []string{"--netmask", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
        netmask: 255.255.255.0
        gateway: 192.168.1.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
        gateway: 192.168.1.1`,
		},
		"unset ipaddr on one network, keep other network": {
			args:    []string{"--netname", "eth0", "--ipaddr", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
      eth0:
        ipaddr: 10.0.0.100
      eth1:
        ipaddr: 172.16.0.100`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
      eth1:
        ipaddr: 172.16.0.100`,
		},

		// Unsetting already unset fields (should be no-op)
		"unset already unset field": {
			args:    []string{"--comment", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset non-existent field with other fields present": {
			args:    []string{"--comment", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    image name: rocky-9`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    image name: rocky-9`,
		},

		// Error cases
		"error: no fields specified": {
			args:    []string{"n01"},
			wantErr: true,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test`,
		},
		"error: non-existent node": {
			args:    []string{"--comment", "nonexistent"},
			wantErr: true,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: test`,
		},

		// Profile inheritance tests
		"unset field from node keeps profile value visible": {
			args:    []string{"--comment", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: profile comment
nodes:
  n01:
    profiles:
    - default
    comment: node comment`,
			outDB: `
nodeprofiles:
  default:
    comment: profile comment
nodes:
  n01:
    profiles:
    - default`,
		},

		// Short flag variations
		"unset with short flags": {
			args:    []string{"-M", "-I", "-G", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
        netmask: 255.255.255.0
        gateway: 192.168.1.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"unset profile with short flag": {
			args:    []string{"-P", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default: {}
nodes:
  n01: {}`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("etc/warewulf/nodes.conf", tt.inDB)
			warewulfd.SetNoDaemon()

			baseCmd := GetCommand()
			args := append(tt.args, "--yes")
			baseCmd.SetArgs(args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err := baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				content := env.ReadFile("etc/warewulf/nodes.conf")
				assert.YAMLEq(t, tt.outDB, content)
			}
		})
	}
}
