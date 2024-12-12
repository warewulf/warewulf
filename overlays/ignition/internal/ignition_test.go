package ignition

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_ignitionOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.ImportFile(t, node.GetNodesConf("etc"), "nodes.conf")
	env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-disks.target.ww", "../rootfs/etc/systemd/system/ww4-disks.target.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww", "../rootfs/etc/systemd/system/ww4-mounts.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/warewulf/ignition.json.ww", "../rootfs/warewulf/ignition.json.ww")

	tests := []struct {
		name string
		args []string
		log  string
		json bool
	}{
		{
			name: "ignition:ww4-disks.target",
			args: []string{"--render", "node1", "ignition", "etc/systemd/system/ww4-disks.target.ww"},
			log:  ignition_disks,
			json: false,
		},
		{
			name: "ignition:ww4-mounts",
			args: []string{"--render", "node1", "ignition", "etc/systemd/system/ww4-mounts.ww"},
			log:  ignition_mounts,
			json: false,
		},
		{
			name: "ignition:ignition.json",
			args: []string{"--quiet", "--render", "node1", "ignition", "warewulf/ignition.json.ww"},
			log:  ignition_json,
			json: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			if tt.json {
				assert.JSONEq(t, tt.log, logbuf.String())
			} else {
				assert.Equal(t, tt.log, logbuf.String())
			}
		})
	}
}

const ignition_disks string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/ww4-disks.target
# This file is autogenerated by warewulf
[Unit]
Description=mount ww4 disks
# make sure that the disks are available
Requires=ignition-ww4-disks.service
After=ignition-ww4-disks.service
Requisite=ignition-ww4-disks.service
# Get the mounts
Wants=scratch.mount
Wants=dev-disk-by\x2dpartlabel-swap.swap
`

const ignition_mounts string = `backupFile: true
writeFile: true
Filename: scratch.mount

# This file is autogenerated by warewulf
[Unit]
ConditionPathExists=/warewulf/ignition.json
Before=local-fs.target
After=ignition-ww4-disks.service
[Mount]
Where=/scratch
What=/dev/disk/by-partlabel/scratch
Type=btrfs
[Install]
RequiredBy=local-fs.target
backupFile: true
writeFile: true
Filename: dev-disk-by\x2dpartlabel-swap.swap
# This file is autogenerated by warewulf
[Unit]
ConditionPathExists=/warewulf/ignition.json
After=ignition-ww4-disks.service
Before=swap.target
[Swap]
What=/dev/disk/by-partlabel/swap
[Install]
RequiredBy=swap.target
`

const ignition_json string = `{
  "ignition": {
    "version": "3.1.0"
  },
  "storage": {
    "disks": [
      {
        "device": "/dev/vda",
        "partitions": [
          {
            "label": "scratch",
            "shouldExist": true,
            "wipePartitionEntry": false
          },
          {
            "label": "swap",
            "number": 1,
            "shouldExist": false,
            "sizeMiB": 1024,
            "wipePartitionEntry": false
          }
        ],
        "wipeTable": true
      }
    ],
    "filesystems": [
      {
        "device": "/dev/disk/by-partlabel/scratch",
        "format": "btrfs",
        "path": "/scratch",
        "wipeFilesystem": true
      },
      {
        "device": "/dev/disk/by-partlabel/swap",
        "format": "swap",
        "path": "swap",
        "wipeFilesystem": false
      }
    ]
  }
}`
