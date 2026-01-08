package unset

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Profile_Unset(t *testing.T) {
	tests := map[string]struct {
		args    []string
		wantErr bool
		inDB    string
		outDB   string
	}{
		// Basic field unsetting
		"unset comment": {
			args:    []string{"--comment", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: test comment
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset cluster": {
			args:    []string{"--cluster", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    cluster name: mycluster
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset image": {
			args:    []string{"--image", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    image name: rockylinux-9
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Network field unsetting
		"unset netmask": {
			args:    []string{"--netmask", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        netmask: 255.255.255.0
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset ipaddr": {
			args:    []string{"--ipaddr", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset gateway": {
			args:    []string{"--gateway", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        gateway: 192.168.1.1
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset hwaddr": {
			args:    []string{"--hwaddr", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        hwaddr: 00:11:22:33:44:55
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset netdev": {
			args:    []string{"--netdev", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset mtu": {
			args:    []string{"--mtu", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        mtu: 9000
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset type": {
			args:    []string{"--type", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        type: ethernet
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Multiple network fields at once
		"unset multiple network fields": {
			args:    []string{"--ipaddr", "--netmask", "--gateway", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
        netmask: 255.255.255.0
        gateway: 192.168.1.1
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// IPv6 fields
		"unset ipaddr6": {
			args:    []string{"--ipaddr6", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr6: fe80::1
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset gateway6": {
			args:    []string{"--gateway6", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        gateway6: fe80::1
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// IPMI fields
		"unset ipmiaddr": {
			args:    []string{"--ipmiaddr", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      ipaddr: 192.168.1.10
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset ipminetmask": {
			args:    []string{"--ipminetmask", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      netmask: 255.255.255.0
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset ipmigateway": {
			args:    []string{"--ipmigateway", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      gateway: 192.168.1.1
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset ipmiuser": {
			args:    []string{"--ipmiuser", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      username: admin
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset ipmipass": {
			args:    []string{"--ipmipass", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      password: secret
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset ipmiport": {
			args:    []string{"--ipmiport", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      port: 623
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Kernel fields
		"unset kernelversion": {
			args:    []string{"--kernelversion", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    kernel:
      version: 5.14.0
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset kernelargs": {
			args:    []string{"--kernelargs", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    kernel:
      args:
      - quiet
      - splash
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Overlay fields
		"unset runtime-overlays": {
			args:    []string{"--runtime-overlays", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    runtime overlay:
      - wwinit
      - generic
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset system-overlays": {
			args:    []string{"--system-overlays", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    system overlay:
      - wwinit
      - generic
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Other fields
		"unset init": {
			args:    []string{"--init", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    init: /sbin/init
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset root": {
			args:    []string{"--root", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    root: initramfs
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset ipxe": {
			args:    []string{"--ipxe", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipxe template: default
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset primarynet": {
			args:    []string{"--primarynet", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    primary network: eth0
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset profiles": {
			args:    []string{"--profile", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    profiles:
      - base
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Multiple profiles
		"unset on multiple profiles": {
			args:    []string{"--comment", "--yes", "default", "compute"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: default profile
  compute:
    comment: compute profile
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
  compute: {}
nodes: {}`,
		},

		// Partial unsetting
		"unset comment but keep image": {
			args:    []string{"--comment", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: test comment
    image name: rockylinux-9
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    image name: rockylinux-9
nodes: {}`,
		},
		"unset netmask but keep ipaddr": {
			args:    []string{"--netmask", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
        netmask: 255.255.255.0
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
nodes: {}`,
		},

		// Network-specific with --netname
		"unset ipaddr on specific network": {
			args:    []string{"--ipaddr", "--netname", "eth1", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
      eth1:
        ipaddr: 10.0.0.100
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
nodes: {}`,
		},

		// Already unset (idempotent)
		"unset already-unset field": {
			args:    []string{"--comment", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Error cases
		"error: no fields specified": {
			args:    []string{"--yes", "default"},
			wantErr: true,
			inDB: `
nodeprofiles:
  default:
    comment: test
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    comment: test
nodes: {}`,
		},
		"error: non-existent profile": {
			args:    []string{"--comment", "--yes", "nonexistent"},
			wantErr: true,
			inDB: `
nodeprofiles:
  default:
    comment: test
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    comment: test
nodes: {}`,
		},
		"force: continue on invalid profile": {
			args:    []string{"--comment", "--yes", "--force", "nonexistent", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: test
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},

		// Multiple fields
		"unset multiple fields": {
			args:    []string{"--comment", "--cluster", "--image", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: test comment
    cluster name: mycluster
    image name: rockylinux-9
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("etc/warewulf/nodes.conf", tt.inDB)
			warewulfd.SetNoDaemon()

			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)

			// Capture output
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)

			err := baseCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Read the output database
			outDB := env.ReadFile("etc/warewulf/nodes.conf")

			// Compare YAML
			assert.YAMLEq(t, tt.outDB, outDB)
		})
	}
}
