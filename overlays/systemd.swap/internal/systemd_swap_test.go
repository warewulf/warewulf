package systemd_swap

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_systemdSwapOverlay(t *testing.T) {
	tests := map[string]struct {
		args      []string
		nodesConf string
		output    string
	}{
		"systemd.swap:disk.swap.ww": {
			args: []string{"--quiet=false", "--render=node1", "systemd.swap", "etc/systemd/system/disk.swap.ww"},
			nodesConf: `
nodes:
  node1:
    filesystems:
      /dev/disk/by-partlabel/rootfs:
        format: ext4
        path: /
      /dev/disk/by-partlabel/scratch:
        format: ext4
        path: /scratch
        wipe_filesystem: true
      /dev/disk/by-partlabel/swap:
        format: swap
        path: swap`,
			output: `backupFile: true
writeFile: true
Filename: dev-disk-by\x2dpartlabel-swap.swap


[Unit]
Before=swap.target

[Swap]
What=/dev/disk/by-partlabel/swap

[Install]
RequiredBy=swap.target
`,
		},

		"systemd.swap:local-fs.target.wants/disk.swap.ww": {
			args: []string{"--quiet=false", "--render=node1", "systemd.swap", "etc/systemd/system/local-fs.target.wants/disk.swap.ww"},
			nodesConf: `
nodes:
  node1:
    filesystems:
      /dev/disk/by-partlabel/rootfs:
        format: ext4
        path: /
      /dev/disk/by-partlabel/scratch:
        format: ext4
        path: /scratch
        wipe_filesystem: true
      /dev/disk/by-partlabel/swap:
        format: swap
        path: swap`,
			output: `backupFile: true
writeFile: true
Filename: dev-disk-by\x2dpartlabel-swap.swap

{{ /* softlink "/etc/systemd/system/dev-disk-by\x2dpartlabel-swap.swap" */ }}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.ImportFile("var/lib/warewulf/overlays/systemd.swap/rootfs/etc/systemd/system/disk.swap.ww", "../rootfs/etc/systemd/system/disk.swap.ww")
			env.ImportFile("var/lib/warewulf/overlays/systemd.swap/rootfs/etc/systemd/system/local-fs.target.wants/disk.swap.ww", "../rootfs/etc/systemd/system/local-fs.target.wants/disk.swap.ww")
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
