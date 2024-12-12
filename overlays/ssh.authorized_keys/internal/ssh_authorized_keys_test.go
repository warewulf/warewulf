package ssh_authorized_keys

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_ssh_authorized_keysOverlay(t *testing.T) {
	t.Skip("ssh.authorized_keys is not yet isolated from the host")

	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.ImportFile(t, node.GetNodesConf("etc"), "nodes.conf")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh.authorized_keys/rootfs/root/.ssh/authorized_keys.ww", "../rootfs/root/.ssh/authorized_keys.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "ssh.authorized_keys:authorized_keys.ww",
			args: []string{"--render", "node1", "ssh.authorized_keys", "root/.ssh/authorized_keys.ww"},
			log:  ssh_authorized_keys,
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

const ssh_authorized_keys string = `backupFile: true
writeFile: true
Filename: root/.ssh/authorized_keys

`
