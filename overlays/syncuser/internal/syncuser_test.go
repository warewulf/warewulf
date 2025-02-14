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
	env := testenv.New(t)
	defer env.RemoveAll()
	env.ImportFile("etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile("var/lib/warewulf/overlays/syncuser/rootfs/etc/passwd.ww", "../rootfs/etc/passwd.ww")
	env.ImportFile("var/lib/warewulf/overlays/syncuser/rootfs/etc/group.ww", "../rootfs/etc/group.ww")
	env.WriteFile("etc/passwd", `
root:x:0:0:root:/root:/bin/bash
user:x:1000:1000:user:/home/user:/bin/bash
`)
	env.WriteFile("etc/group", `
root:x:0:
user:x:1000:
`)
	env.WriteFile("var/lib/warewulf/chroots/rockylinux-9/rootfs/etc/passwd", `
root:x:0:0:root:/root:/bin/bash
`)
	env.WriteFile("var/lib/warewulf/chroots/rockylinux-9/rootfs/etc/group", `
root:x:0:
`)

	tests := map[string]struct {
		args []string
		log  string
	}{
		"syncuser:passwd.ww": {
			args: []string{"--render", "node1", "syncuser", "etc/passwd.ww"},
			log: `backupFile: true
writeFile: true
Filename: etc/passwd
root:x:0:0:root:/root:/bin/bash
user:x:1000:1000:user:/home/user:/bin/bash
`,
		},
		"syncuser:passwd.ww (passwordless root)": {
			args: []string{"--render", "node2", "syncuser", "etc/passwd.ww"},
			log: `backupFile: true
writeFile: true
Filename: etc/passwd
root::0:0:root:/root:/bin/bash
user:x:1000:1000:user:/home/user:/bin/bash
`,
		},
		"syncuser:group.ww": {
			args: []string{"--render", "node1", "syncuser", "etc/group.ww"},
			log: `backupFile: true
writeFile: true
Filename: etc/group
root:x:0:
user:x:1000:
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
