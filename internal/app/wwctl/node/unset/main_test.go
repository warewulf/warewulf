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

		// Tag deletion tests
		"delete node tag": {
			args:    []string{"--tag=mytag", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      mytag: myvalue
      keeptag: keepvalue`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      keeptag: keepvalue`,
		},
		"delete multiple node tags": {
			args:    []string{"--tag=tag1,tag2", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      tag1: val1
      tag2: val2
      tag3: val3`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      tag3: val3`,
		},
		"delete all node tags leaves empty node": {
			args:    []string{"--tag=only", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      only: value`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"delete nonexistent tag is noop": {
			args:    []string{"--tag=nonexistent", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      keep: value`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      keep: value`,
		},
		"delete tag from node with no tags": {
			args:    []string{"--tag=anytag", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: hello`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: hello`,
		},
		"delete ipmi tag": {
			args:    []string{"--ipmitag=bmctag", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      username: admin
      tags:
        bmctag: bmcvalue`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      username: admin`,
		},
		"delete ipmi tag when no ipmi section": {
			args:    []string{"--ipmitag=anytag", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: hello`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: hello`,
		},
		"delete net tag on default network": {
			args:    []string{"--nettag=dns1", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
        tags:
          dns1: 8.8.8.8
          dns2: 8.8.4.4`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
        tags:
          dns2: 8.8.4.4`,
		},
		"delete net tag on specific network": {
			args:    []string{"--netname", "eth1", "--nettag=mytag", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        tags:
          mytag: defaultval
      eth1:
        ipaddr: 10.0.0.1
        tags:
          mytag: eth1val`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        tags:
          mytag: defaultval
      eth1:
        ipaddr: 10.0.0.1`,
		},
		"delete net tag on nonexistent network is noop": {
			args:    []string{"--netname", "nonet", "--nettag=anytag", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.1`,
		},
		"combine tag deletion with field unset": {
			args:    []string{"--comment", "--tag=mytag", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: hello
    tags:
      mytag: val
      keep: val2`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    tags:
      keep: val2`,
		},

		// Object deletion tests
		"unset net removes entire netdev": {
			args:    []string{"--net=eth0", "n01"},
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
		"unset net removes only named netdev": {
			args:    []string{"--net=eth0,eth1", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100
      eth0:
        ipaddr: 10.0.0.1
      eth1:
        ipaddr: 10.0.0.2`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100`,
		},
		"unset net nonexistent is noop": {
			args:    []string{"--net=nonet", "n01"},
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
  n01:
    network devices:
      default:
        ipaddr: 192.168.1.100`,
		},
		"unset disk removes entire disk": {
			args:    []string{"--disk=/dev/vda", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
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
            size_mib: "102400"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vdb:
        partitions:
          data:
            number: "1"
            size_mib: "102400"`,
		},
		"unset disk nonexistent is noop": {
			args:    []string{"--disk=/dev/vdz", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"`,
		},
		"unset part removes partition from disk": {
			args:    []string{"--part=swap", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"
          root:
            number: "2"
            size_mib: "51200"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "2"
            size_mib: "51200"`,
		},
		"unset part nonexistent is noop": {
			args:    []string{"--part=nopart", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          swap:
            number: "1"
            size_mib: "4096"`,
		},
		"unset fs removes entire filesystem": {
			args:    []string{"--fs=/dev/disk/by-partlabel/swap", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/swap:
        format: swap
      /dev/disk/by-partlabel/root:
        format: ext4`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4`,
		},
		"unset fs nonexistent is noop": {
			args:    []string{"--fs=/dev/nofs", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4`,
		},

		// Scoped unset: disk/partition/filesystem
		"scoped partnumber clears only targeted partition": {
			args:    []string{"--partnumber", "--diskname=/dev/vda", "--partname=root", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
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
            size_mib: "2048"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
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
            size_mib: "2048"`,
		},
		"scoped diskwipe clears only targeted disk": {
			args:    []string{"--diskwipe", "--diskname=/dev/vda", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        wipe_table: true
        partitions:
          root:
            number: "1"
      /dev/vdb:
        wipe_table: true
        partitions:
          data:
            number: "1"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "1"
      /dev/vdb:
        wipe_table: true
        partitions:
          data:
            number: "1"`,
		},
		"scoped fsformat clears only targeted filesystem": {
			args:    []string{"--fsformat", "--fsname=/dev/disk/by-partlabel/root", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4
        path: /
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/root:
        path: /
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var`,
		},
		"error: partnumber without diskname/partname": {
			args:    []string{"--partnumber", "n01"},
			wantErr: true,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "1"
            size_mib: "1024"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          root:
            number: "1"
            size_mib: "1024"`,
		},
		"error: diskwipe without diskname": {
			args:    []string{"--diskwipe", "n01"},
			wantErr: true,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        wipe_table: true`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        wipe_table: true`,
		},
		"error: fsformat without fsname": {
			args:    []string{"--fsformat", "n01"},
			wantErr: true,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    filesystems:
      /dev/disk/by-partlabel/root:
        format: ext4`,
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
			}
			content := env.ReadFile("etc/warewulf/nodes.conf")
			assert.YAMLEq(t, tt.outDB, content)
		})
	}
}
