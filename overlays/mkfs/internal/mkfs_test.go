package mkfs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_mkfsOverlay(t *testing.T) {
	tests := map[string]struct {
		args      []string
		nodesConf string
		output    string
	}{
		"mkfs:20-mkfs.sh.ww (empty)": {
			args: []string{"--quiet=true", "--render=node1", "mkfs", "warewulf/wwinit.d/20-mkfs.sh.ww"},
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

already_formatted() {
    if ! command -v wipefs >/dev/null ; then
        info "warewulf: wipefs not found, cannot check if device is already formatted"
        return 0
    fi

    if wipefs -n "${1}" &>/dev/null; then
        info "warewulf: ${1} already formatted"
        return 0
    fi

    return 1
}

if command -v mkfs >/dev/null ; then :
else
    info "warewulf: mkfs not found"
fi
`,
		},

		"mkfs:20-mkfs.sh.ww (resource)": {
			args: []string{"--quiet=true", "--render=node1", "mkfs", "warewulf/wwinit.d/20-mkfs.sh.ww"},
			nodesConf: `
nodes:
  node1:
    resources:
      mkfs:
        - device: /dev/disk/by-partlabel/rootfs
          type: ext4
        - device: /dev/disk/by-partlabel/scratch
          type: ext4
          overwrite: true`,
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

already_formatted() {
    if ! command -v wipefs >/dev/null ; then
        info "warewulf: wipefs not found, cannot check if device is already formatted"
        return 0
    fi

    if wipefs -n "${1}" &>/dev/null; then
        info "warewulf: ${1} already formatted"
        return 0
    fi

    return 1
}

if command -v mkfs >/dev/null ; then :
    if false || ! already_formatted /dev/disk/by-partlabel/rootfs; then
        info "warewulf: mkfs: formatting /dev/disk/by-partlabel/rootfs"
        mkfs --type=ext4    /dev/disk/by-partlabel/rootfs  || die "warewulf: mkfs: failed to format /dev/disk/by-partlabel/rootfs"
    else
        info "warewulf: mkfs: skipping /dev/disk/by-partlabel/rootfs"
    fi
    if true || ! already_formatted /dev/disk/by-partlabel/scratch; then
        info "warewulf: mkfs: formatting /dev/disk/by-partlabel/scratch"
        mkfs --type=ext4   -F /dev/disk/by-partlabel/scratch  || die "warewulf: mkfs: failed to format /dev/disk/by-partlabel/scratch"
    else
        info "warewulf: mkfs: skipping /dev/disk/by-partlabel/scratch"
    fi
else
    info "warewulf: mkfs not found"
fi
`,
		},

		"mkfs:20-mkfs.sh.ww (native)": {
			args: []string{"--quiet=true", "--render=node1", "mkfs", "warewulf/wwinit.d/20-mkfs.sh.ww"},
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

already_formatted() {
    if ! command -v wipefs >/dev/null ; then
        info "warewulf: wipefs not found, cannot check if device is already formatted"
        return 0
    fi

    if wipefs -n "${1}" &>/dev/null; then
        info "warewulf: ${1} already formatted"
        return 0
    fi

    return 1
}

if command -v mkfs >/dev/null ; then :
    if false || ! already_formatted /dev/disk/by-partlabel/rootfs; then
        info "warewulf: mkfs: formatting /dev/disk/by-partlabel/rootfs"
        mkfs --type=ext4    /dev/disk/by-partlabel/rootfs  || die "warewulf: mkfs: failed to format /dev/disk/by-partlabel/rootfs"
    else
        info "warewulf: mkfs: skipping /dev/disk/by-partlabel/rootfs"
    fi
    if true || ! already_formatted /dev/disk/by-partlabel/scratch; then
        info "warewulf: mkfs: formatting /dev/disk/by-partlabel/scratch"
        mkfs --type=ext4    -F /dev/disk/by-partlabel/scratch  || die "warewulf: mkfs: failed to format /dev/disk/by-partlabel/scratch"
    else
        info "warewulf: mkfs: skipping /dev/disk/by-partlabel/scratch"
    fi
else
    info "warewulf: mkfs not found"
fi
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.ImportFile("var/lib/warewulf/overlays/mkfs/rootfs/warewulf/wwinit.d/20-mkfs.sh.ww", "../rootfs/warewulf/wwinit.d/20-mkfs.sh.ww")
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
