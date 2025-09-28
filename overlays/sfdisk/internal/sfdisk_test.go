package sfdisk

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_sfdiskOverlay(t *testing.T) {
	tests := map[string]struct {
		args      []string
		nodesConf string
		output    string
	}{
		"sfdisk:disks.ww (empty)": {
			args: []string{"--quiet=false", "--render=node1", "sfdisk", "warewulf/sfdisk/disks.ww"},
			nodesConf: `
nodes:
  node1: {}`,
			output: `backupFile: true
writeFile: true
Filename: warewulf/sfdisk/disks

`,
		},
		"sfdisk:disks.ww (resource)": {
			args: []string{"--quiet=false", "--render=node1", "sfdisk", "warewulf/sfdisk/disks.ww"},
			nodesConf: `
nodes:
  node1:
    resources:
      sfdisk:
        - device: /dev/sda
          first-lba: 34
          label: gpt
          last-lba: 20971486
          partitions:
            - device: /dev/sda1
              name: sfdisk-rootfs
              size: 4194304
              start: 2048
              type: linux
            - device: /dev/sda2
              name: sfdisk-scratch
              size: 1048576
              start: 4196352
              type: linux
            - device: /dev/sda3
              name: sfdisk-swap
              size: 2097152
              start: 5244928
              type: linux
          sector-size: 512`,
			output: `backupFile: true
writeFile: true
Filename: device-0

label: gpt
first-lba: 34
last-lba: 20971486
sector-size: 512
/dev/sda1 : start=2048 size=4194304 name=sfdisk-rootfs type=linux
/dev/sda2 : start=4196352 size=1048576 name=sfdisk-scratch type=linux
/dev/sda3 : start=5244928 size=2097152 name=sfdisk-swap type=linux
`,
		},
		"sfdisk:disks.ww (native)": {
			args: []string{"--quiet=false", "--render=node1", "sfdisk", "warewulf/sfdisk/disks.ww"},
			nodesConf: `
nodes:
  node1:
    disks:
      /dev/sda:
        wipe_table: true
        partitions:
          rootfs:
            number: "1"
            start_mib: "2"
            size_mib: "4096"
            should_exist: true
            type_guid: linux
          scratch:
            number: "2"
            start_mib: "4098"
            size_mib: "10240"
            should_exist: true
            type_guid: linux
          swap:
            number: "3"
            start_mib: "5122"
            size_mib: "2048"
            should_exist: true
            type_guid: swap`,
			output: `backupFile: true
writeFile: true
Filename: device-0

start=2MiB size=4096MiB name=rootfs type=linux
start=4098MiB size=10240MiB name=scratch type=linux
start=5122MiB size=2048MiB name=swap type=swap
`,
		},
		"sfdisk:10-sfdisk.sh.ww (empty)": {
			args: []string{"--quiet", "--render=node1", "sfdisk", "warewulf/wwinit.d/10-sfdisk.sh.ww"},
			nodesConf: `
nodes:
  node1: {}`,
			output: `#!/bin/sh

PATH=$PATH:/sbin:/usr/sbin:/bin:/usr/bin

if ! command -v info >/dev/null; then
    info() {
        printf '%s\n' "$*"
    }
fi

if ! command -v die >/dev/null; then
    die() {
        printf '%s\n' "$*" >&2
        exit 1
    }
fi

if ! command -v sfdisk >/dev/null ; then
    info "warewulf: sfdisk not found, skipping partitioning"
else :
fi
`,
		},
		"sfdisk:10-sfdisk.sh.ww (resource)": {
			args: []string{"--quiet", "--render=node1", "sfdisk", "warewulf/wwinit.d/10-sfdisk.sh.ww"},
			nodesConf: `
nodes:
  node1: 
    resources:
      sfdisk:
        - device: /dev/sda
          first-lba: 34
          label: gpt
          last-lba: 20971486
          partitions:
            - device: /dev/sda1
              name: sfdisk-rootfs
              size: 4194304
              start: 2048
              type: linux
            - device: /dev/sda2
              name: sfdisk-scratch
              size: 1048576
              start: 4196352
              type: linux
            - device: /dev/sda3
              name: sfdisk-swap
              size: 2097152
              start: 5244928
              type: linux
          sector-size: 512`,
			output: `#!/bin/sh

PATH=$PATH:/sbin:/usr/sbin:/bin:/usr/bin

if ! command -v info >/dev/null; then
    info() {
        printf '%s\n' "$*"
    }
fi

if ! command -v die >/dev/null; then
    die() {
        printf '%s\n' "$*" >&2
        exit 1
    }
fi

if ! command -v sfdisk >/dev/null ; then
    info "warewulf: sfdisk not found, skipping partitioning"
else :
    info "warewulf: sfdisk: partitioning /dev/sda"
    sfdisk --label gpt --wipe "auto" "/dev/sda" < "${PREFIX}/warewulf/sfdisk/device-0" || die "warewulf: sfdisk: failed to partition /dev/sda"

    if command -v blockdev >/dev/null ; then
        info "warewulf: blockdev: re-reading partition table"
        blockdev --rereadpt /dev/sda
    fi
    if command -v udevadm >/dev/null ; then
        info "warewulf: udevadm: triggering udev events for block devices"
        udevadm trigger --subsystem-match=block --action=add
        udevadm settle
    fi
fi
`,
		},
		"sfdisk:10-sfdisk.sh.ww (native)": {
			args: []string{"--quiet", "--render=node1", "sfdisk", "warewulf/wwinit.d/10-sfdisk.sh.ww"},
			nodesConf: `
nodes:
  node1:
    disks:
      /dev/sda:
        partitions:
          rootfs:
            number: "1"
            size_mib: "4096"
            should_exist: true
          scratch:
            number: "2"
            size_mib: "10240"
            should_exist: true
          swap:
            number: "3"
            size_mib: "2048"
            should_exist: true
      /dev/sdb:
        wipe_table: true`,
			output: `#!/bin/sh

PATH=$PATH:/sbin:/usr/sbin:/bin:/usr/bin

if ! command -v info >/dev/null; then
    info() {
        printf '%s\n' "$*"
    }
fi

if ! command -v die >/dev/null; then
    die() {
        printf '%s\n' "$*" >&2
        exit 1
    }
fi

if ! command -v sfdisk >/dev/null ; then
    info "warewulf: sfdisk not found, skipping partitioning"
else :
    info "warewulf: sfdisk: partitioning /dev/sda"
    sfdisk --label gpt --wipe "auto" "/dev/sda" < "${PREFIX}/warewulf/sfdisk/device-0" || die "warewulf: sfdisk: failed to partition /dev/sda"

    if command -v blockdev >/dev/null ; then
        info "warewulf: blockdev: re-reading partition table"
        blockdev --rereadpt /dev/sda
    fi
    info "warewulf: sfdisk: partitioning /dev/sdb"
    sfdisk --label gpt --wipe "always" "/dev/sdb" < "${PREFIX}/warewulf/sfdisk/device-1" || die "warewulf: sfdisk: failed to partition /dev/sdb"

    if command -v blockdev >/dev/null ; then
        info "warewulf: blockdev: re-reading partition table"
        blockdev --rereadpt /dev/sdb
    fi
    if command -v udevadm >/dev/null ; then
        info "warewulf: udevadm: triggering udev events for block devices"
        udevadm trigger --subsystem-match=block --action=add
        udevadm settle
    fi
fi
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.ImportFile("var/lib/warewulf/overlays/sfdisk/rootfs/warewulf/sfdisk/disks.ww", "../rootfs/warewulf/sfdisk/disks.ww")
			env.ImportFile("var/lib/warewulf/overlays/sfdisk/rootfs/warewulf/wwinit.d/10-sfdisk.sh.ww", "../rootfs/warewulf/wwinit.d/10-sfdisk.sh.ww")
			env.WriteFile("etc/warewulf/nodes.conf", tt.nodesConf)
			cmd := show.GetCommand()
			cmd.SetArgs(tt.args)
			stdout := bytes.NewBufferString("")
			stderr := bytes.NewBufferString("")
			logbuf := bytes.NewBufferString("")
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			wwlog.SetLogWriter(logbuf)
			err := cmd.Execute()
			assert.NoError(t, err)
			assert.Empty(t, stdout.String())
			assert.Empty(t, stderr.String())
			assert.Equal(t, tt.output, logbuf.String())
		})
	}
}
