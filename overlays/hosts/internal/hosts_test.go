package hosts

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_hostsOverlay(t *testing.T) {
	hostname, _ := os.Hostname()
	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.ImportFile(t, "etc/warewulf/warewulf.conf", "warewulf.conf")
	assert.NoError(t, config.Get().Read(env.GetPath("etc/warewulf/warewulf.conf")))
	env.ImportFile(t, node.GetNodesConf("etc"), "nodes.conf")
	env.ImportFile(t, "var/lib/warewulf/overlays/hosts/rootfs/etc/hosts.ww", "../rootfs/etc/hosts.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "/etc/hosts",
			args: []string{"--render", "node1", "hosts", "etc/hosts.ww"},
			log:  hosts,
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
			assert.Equal(t, strings.Replace(tt.log, "%HOSTNAME%", hostname, -1), logbuf.String())
		})
	}
}

const hosts string = `backupFile: true
writeFile: true
Filename: etc/hosts
127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4
::1         localhost localhost.localdomain localhost6 localhost6.localdomain6

# Warewulf Server
192.168.0.1 %HOSTNAME% warewulf
# Entry for node1
192.168.3.21 node1 node1-default node1-wwnet0
192.168.3.22  node1-secondary node1-wwnet1
# Entry for node2
192.168.3.23 node2 node2-default node2-wwnet0
`
