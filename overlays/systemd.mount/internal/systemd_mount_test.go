package systemd_mount

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_systemdMountOverlay(t *testing.T) {
	tests := map[string]struct {
		args      []string
		nodesConf string
		output    string
	}{
		"systemd.mount:disk.mount.ww": {
			args: []string{"--quiet=false", "--render=node1", "systemd.mount", "etc/systemd/system/disk.mount.ww"},
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
Filename: -.mount


[Unit]
Before=local-fs.target

[Mount]
Where=/
What=/dev/disk/by-partlabel/rootfs
Type=ext4

[Install]
RequiredBy=local-fs.target
backupFile: true
writeFile: true
Filename: scratch.mount

[Unit]
Before=local-fs.target

[Mount]
Where=/scratch
What=/dev/disk/by-partlabel/scratch
Type=ext4

[Install]
RequiredBy=local-fs.target
`,
		},

		"systemd.mount:local-fs.target.wants/disk.mount.ww": {
			args: []string{"--quiet=false", "--render=node1", "systemd.mount", "etc/systemd/system/local-fs.target.wants/disk.mount.ww"},
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
Filename: -.mount

{{ /* softlink "/etc/systemd/system/-.mount" */ }}
backupFile: true
writeFile: true
Filename: scratch.mount
{{ /* softlink "/etc/systemd/system/scratch.mount" */ }}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.ImportFile("var/lib/warewulf/overlays/systemd.mount/rootfs/etc/systemd/system/disk.mount.ww", "../rootfs/etc/systemd/system/disk.mount.ww")
			env.ImportFile("var/lib/warewulf/overlays/systemd.mount/rootfs/etc/systemd/system/local-fs.target.wants/disk.mount.ww", "../rootfs/etc/systemd/system/local-fs.target.wants/disk.mount.ww")
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
