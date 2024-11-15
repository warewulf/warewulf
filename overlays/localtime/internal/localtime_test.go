package localtime

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_localtimeOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.ImportFile(t, "etc/warewulf/nodes.conf", "nodes.conf")
	assert.NoError(t, config.Get().Read(env.GetPath("etc/warewulf/warewulf.conf")))
	env.ImportFile(t, "var/lib/warewulf/overlays/localtime/rootfs/etc/localtime.ww", "../rootfs/etc/localtime.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "/etc/localtime",
			args: []string{"--render", "node1", "localtime", "etc/localtime.ww"},
			log:  localtime,
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

const localtime string = `backupFile: true
writeFile: true
Filename: etc/localtime
{{ /* softlink "/usr/share/zoneinfo/GMT" */ }}
`
