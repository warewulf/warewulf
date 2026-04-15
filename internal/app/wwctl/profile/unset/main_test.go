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

		// Tag deletion tests
		"delete profile tag": {
			args:    []string{"--tag=mytag", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    tags:
      mytag: myvalue
      keeptag: keepvalue
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    tags:
      keeptag: keepvalue
nodes: {}`,
		},
		"delete multiple profile tags": {
			args:    []string{"--tag=tag1,tag2", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    tags:
      tag1: val1
      tag2: val2
      tag3: val3
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    tags:
      tag3: val3
nodes: {}`,
		},
		"delete ipmi tag from profile": {
			args:    []string{"--ipmitag=bmctag", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      username: admin
      tags:
        bmctag: bmcvalue
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    ipmi:
      username: admin
nodes: {}`,
		},
		"delete net tag from profile": {
			args:    []string{"--nettag=dns1", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
        tags:
          dns1: 8.8.8.8
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
nodes: {}`,
		},
		"delete net tag on specific network from profile": {
			args:    []string{"--netname", "eth1", "--nettag=mytag", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        tags:
          mytag: defaultval
      eth1:
        ipaddr: 10.0.0.1
        tags:
          mytag: eth1val
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    network devices:
      default:
        tags:
          mytag: defaultval
      eth1:
        ipaddr: 10.0.0.1
nodes: {}`,
		},
		"combine tag deletion with field unset on profile": {
			args:    []string{"--comment", "--tag=mytag", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: hello
    tags:
      mytag: val
      keep: val2
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    tags:
      keep: val2
nodes: {}`,
		},

		// Object deletion tests
		"unset net removes entire netdev from profile": {
			args:    []string{"--net=eth0", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
      eth0:
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
		"unset net nonexistent is noop on profile": {
			args:    []string{"--net=nonet", "--yes", "default"},
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
  default:
    network devices:
      default:
        ipaddr: 192.168.1.100
nodes: {}`,
		},
		"unset disk removes entire disk from profile": {
			args:    []string{"--disk=/dev/vda", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
      /dev/vdb:
        partitions:
          data:
            number: "1"
            size_mib: "102400"
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    disks:
      /dev/vdb:
        partitions:
          data:
            number: "1"
            size_mib: "102400"
nodes: {}`,
		},
		"unset part removes partition from profile disk": {
			args:    []string{"--part=swap", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
          root:
            number: "2"
            size_mib: "51200"
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "2"
            size_mib: "51200"
nodes: {}`,
		},
		"unset part scoped to diskname removes partition only from that disk": {
			args:    []string{"--part=swap", "--diskname=/dev/vda", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
      /dev/vdb:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    disks:
      /dev/vdb:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
nodes: {}`,
		},
		"unset part unscoped removes partition from all disks": {
			args:    []string{"--part=swap", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
      /dev/vdb:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
nodes: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"unset fs removes filesystem from profile": {
			args:    []string{"--fs=/dev/disk/by-partlabel/swap", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    filesystems:
      /dev/disk/by-partlabel/swap:
        format: swap
      /dev/disk/by-partlabel/root:
        format: ext4
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4
nodes: {}`,
		},

		// Scoped unset: disk/partition/filesystem
		"scoped partnumber clears only targeted partition": {
			args:    []string{"--partnumber", "--diskname=/dev/vda", "--partname=root", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "1"
            size_mib: "1024"
          swap:
            number: "2"
            size_mib: "512"
      /dev/vdb:
        partitions:
          data:
            number: "1"
            size_mib: "2048"
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          root:
            size_mib: "1024"
          swap:
            number: "2"
            size_mib: "512"
      /dev/vdb:
        partitions:
          data:
            number: "1"
            size_mib: "2048"
nodes: {}`,
		},
		"scoped fsformat clears only targeted filesystem": {
			args:    []string{"--fsformat", "--fsname=/dev/disk/by-partlabel/root", "--yes", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4
        path: /
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    filesystems:
      /dev/disk/by-partlabel/root:
        path: /
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}`,
		},
		"error: partnumber without diskname/partname": {
			args:    []string{"--partnumber", "--yes", "default"},
			wantErr: true,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "1"
            size_mib: "1024"
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "1"
            size_mib: "1024"
nodes: {}`,
		},
		"error: fsformat without fsname": {
			args:    []string{"--fsformat", "--yes", "default"},
			wantErr: true,
			inDB: `
nodeprofiles:
  default:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4
nodes: {}`,
			outDB: `
nodeprofiles:
  default:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4
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
