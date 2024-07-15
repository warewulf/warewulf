package syncuser

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_syncuserOverlay(t *testing.T) {
	t.Skip("syncuser is not yet isolated from the host")

	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.ImportFile(t, "etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile(t, "var/lib/warewulf/overlays/syncuser/rootfs/etc/passwd.ww", "../../../../../overlays/syncuser/rootfs/etc/passwd.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/syncuser/rootfs/etc/group.ww", "../../../../../overlays/syncuser/rootfs/etc/group.ww")
	env.WriteFile(t, "var/lib/warewulf/chroots/rockylinux-9/rootfs/etc/passwd", `root:x:0:0:root:/root:/bin/bash`)
	env.WriteFile(t, "var/lib/warewulf/chroots/rockylinux-9/rootfs/etc/group", `root:x:0:`)

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "syncuser:passwd.ww",
			args: []string{"--render", "node1", "syncuser", "etc/passwd.ww"},
			log:  syncuser_passwd,
		},
		{
			name: "syncuser:group.ww",
			args: []string{"--render", "node1", "syncuser", "etc/group.ww"},
			log:  syncuser_group,
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
			assert.Equal(t, tt.log, logbuf.String())
		})
	}
}

const syncuser_passwd string = `backupFile: true
writeFile: true
Filename: etc/passwd
# Uncomment the following line to enable passwordless root login
# root::0:0:root:/root:/bin/bash
root:x:0:0:root:/root:/bin/bash
`

const syncuser_group string = `backupFile: true
writeFile: true
Filename: etc/group
root:x:0:
`
