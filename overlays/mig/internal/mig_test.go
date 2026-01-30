package mig

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_migOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.ImportFile("etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile("var/lib/warewulf/overlays/mig/rootfs/etc/systemd/system/nvidia-mig.service.ww", "../rootfs/etc/systemd/system/nvidia-mig.service.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "mig:nvidia-mig.service (default)",
			args: []string{"--render", "node1", "mig", "etc/systemd/system/nvidia-mig.service.ww"},
			log:  migServiceDefault,
		},
		{
			name: "mig:nvidia-mig.service (per-gpu)",
			args: []string{"--render", "node2", "mig", "etc/systemd/system/nvidia-mig.service.ww"},
			log:  migServicePerGpu,
		},
		{
			name: "mig:nvidia-mig.service (custom)",
			args: []string{"--render", "node3", "mig", "etc/systemd/system/nvidia-mig.service.ww"},
			log:  migServiceCustom,
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

const migServiceDefault string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/nvidia-mig.service
[Unit]
DefaultDependencies=no
Description=Configure NVIDIA MIG (Multi-Instance GPU) partitions at boot time
Before=nvidia-persistenced.service nvidia-dcgm.service dcgm_exporter.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/local/sbin/_config_mig "2,2"
ExecStop=sh -c 'nvidia-smi mig -dci ; nvidia-smi mig -dgi ; nvidia-smi -mig 0'

[Install]
WantedBy=multi-user.target
`

const migServicePerGpu string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/nvidia-mig.service
[Unit]
DefaultDependencies=no
Description=Configure NVIDIA MIG (Multi-Instance GPU) partitions at boot time
Before=nvidia-persistenced.service nvidia-dcgm.service dcgm_exporter.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/local/sbin/_config_mig "14,14,14,15 7,7,7,8"
ExecStop=sh -c 'nvidia-smi mig -dci ; nvidia-smi mig -dgi ; nvidia-smi -mig 0'

[Install]
WantedBy=multi-user.target
`

const migServiceCustom string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/nvidia-mig.service
[Unit]
DefaultDependencies=no
Description=Configure NVIDIA MIG (Multi-Instance GPU) partitions at boot time
Before=nvidia-persistenced.service nvidia-dcgm.service dcgm_exporter.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/local/sbin/_config_mig "1,1,1,1"
ExecStop=sh -c 'nvidia-smi mig -dci ; nvidia-smi mig -dgi ; nvidia-smi -mig 0'

[Install]
WantedBy=multi-user.target
`
