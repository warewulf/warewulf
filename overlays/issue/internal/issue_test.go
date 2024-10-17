package issue

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_issueOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.ImportFile(t, "etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile(t, "var/lib/warewulf/overlays/issue/rootfs/etc/issue.ww", "../rootfs/etc/issue.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "/etc/issue",
			args: []string{"--render", "node1", "issue", "etc/issue.ww"},
			log:  issue,
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

const issue string = `backupFile: true
writeFile: true
Filename: etc/issue
Warewulf Node:      node1
Container:          rockylinux-9
Kernelargs:         quiet crashkernel=no vga=791 net.naming-scheme=v238

Network:
    default: wwnet0
    default: \4{wwnet0} (configured: 192.168.3.21)
    default: e6:92:39:49:7b:03
    secondary: wwnet1
    secondary: \4{wwnet1} (configured: 192.168.3.22)
    secondary: 9a:77:29:73:14:f1
`
