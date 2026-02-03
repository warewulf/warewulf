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
	env.ImportFile("var/lib/warewulf/overlays/mig/rootfs/etc/systemd/system/ww-nvidia-mig.service.ww", "../rootfs/etc/systemd/system/ww-nvidia-mig.service.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "mig:nvidia-mig.service (default)",
			args: []string{"--render", "node1", "mig", "etc/systemd/system/ww-nvidia-mig.service.ww"},
			log:  migServiceDefault,
		},
		{
			name: "mig:nvidia-mig.service (homogeneous)",
			args: []string{"--render", "node2", "mig", "etc/systemd/system/ww-nvidia-mig.service.ww"},
			log:  migServiceHomogeneous,
		},
		{
			name: "mig:nvidia-mig.service (heterogeneous)",
			args: []string{"--render", "node3", "mig", "etc/systemd/system/ww-nvidia-mig.service.ww"},
			log:  migServiceHeterogeneous,
		},
		{
			name: "mig:nvidia-mig.service (heterogeneous short names)",
			args: []string{"--render", "node4", "mig", "etc/systemd/system/ww-nvidia-mig.service.ww"},
			log:  migServiceHeterogeneousShortNames,
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
Filename: etc/systemd/system/ww-nvidia-mig.service
[Unit]
DefaultDependencies=no
Description=Configure NVIDIA MIG (Multi-Instance GPU) partitions via Warewulf
Before=nvidia-persistenced.service nvidia-dcgm.service dcgm_exporter.service nvsm.service nvsm-core.service
ConditionPathExists=/usr/bin/nvidia-smi

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStartPre=-/usr/bin/nvidia-smi -pm 1
ExecStartPre=/usr/bin/nvidia-smi -mig 1
ExecStart=/usr/bin/nvidia-smi mig -cgi 2,2 -C
ExecStartPost=-/usr/bin/nvidia-smi -L
ExecStartPost=-/usr/bin/sh /usr/local/sbin/mig2gres
ExecStop=-/usr/bin/nvidia-smi mig -dci
ExecStop=-/usr/bin/nvidia-smi mig -dgi
ExecStop=/usr/bin/nvidia-smi -mig 0
TimeoutStartSec=300
TimeoutStopSec=300

[Install]
WantedBy=multi-user.target
`

const migServiceHomogeneous string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/ww-nvidia-mig.service
[Unit]
DefaultDependencies=no
Description=Configure NVIDIA MIG (Multi-Instance GPU) partitions via Warewulf
Before=nvidia-persistenced.service nvidia-dcgm.service dcgm_exporter.service nvsm.service nvsm-core.service
ConditionPathExists=/usr/bin/nvidia-smi

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStartPre=-/usr/bin/nvidia-smi -pm 1
ExecStartPre=/usr/bin/nvidia-smi -mig 1
ExecStart=/usr/bin/nvidia-smi mig -cgi 14,14,14,15 -C
ExecStartPost=-/usr/bin/nvidia-smi -L
ExecStartPost=-/usr/bin/sh /usr/local/sbin/mig2gres
ExecStop=-/usr/bin/nvidia-smi mig -dci
ExecStop=-/usr/bin/nvidia-smi mig -dgi
ExecStop=/usr/bin/nvidia-smi -mig 0
TimeoutStartSec=300
TimeoutStopSec=300

[Install]
WantedBy=multi-user.target
`

const migServiceHeterogeneous string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/ww-nvidia-mig.service
[Unit]
DefaultDependencies=no
Description=Configure NVIDIA MIG (Multi-Instance GPU) partitions via Warewulf
Before=nvidia-persistenced.service nvidia-dcgm.service dcgm_exporter.service nvsm.service nvsm-core.service
ConditionPathExists=/usr/bin/nvidia-smi

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStartPre=-/usr/bin/nvidia-smi -pm 1
ExecStartPre=/usr/bin/nvidia-smi -mig 1
ExecStart=/usr/bin/nvidia-smi mig -cgi 14,14,14,15 -i 0 -C
ExecStart=/usr/bin/nvidia-smi mig -cgi 9,9 -i 1 -C
ExecStart=/usr/bin/nvidia-smi mig -cgi 0 -i 2 -C
ExecStart=/usr/bin/nvidia-smi mig -cgi 0 -i 3 -C
ExecStartPost=-/usr/bin/nvidia-smi -L
ExecStartPost=-/usr/bin/sh /usr/local/sbin/mig2gres
ExecStop=-/usr/bin/nvidia-smi mig -dci
ExecStop=-/usr/bin/nvidia-smi mig -dgi
ExecStop=/usr/bin/nvidia-smi -mig 0
TimeoutStartSec=300
TimeoutStopSec=300

[Install]
WantedBy=multi-user.target
`

const migServiceHeterogeneousShortNames string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/ww-nvidia-mig.service
[Unit]
DefaultDependencies=no
Description=Configure NVIDIA MIG (Multi-Instance GPU) partitions via Warewulf
Before=nvidia-persistenced.service nvidia-dcgm.service dcgm_exporter.service nvsm.service nvsm-core.service
ConditionPathExists=/usr/bin/nvidia-smi

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStartPre=-/usr/bin/nvidia-smi -pm 1
ExecStartPre=/usr/bin/nvidia-smi -mig 1
ExecStart=/usr/bin/nvidia-smi mig -cgi 1g.23gb+me,1g.23gb,1g.23gb,1g.23gb,1g.23gb,1g.23gb,1g.23gb -i 0 -C
ExecStart=/usr/bin/nvidia-smi mig -cgi 3g.90gb,3g.90gb -i 1 -C
ExecStart=/usr/bin/nvidia-smi mig -cgi 0 -i 2 -C
ExecStart=/usr/bin/nvidia-smi mig -cgi 0 -i 3 -C
ExecStartPost=-/usr/bin/nvidia-smi -L
ExecStartPost=-/usr/bin/sh /usr/local/sbin/mig2gres
ExecStop=-/usr/bin/nvidia-smi mig -dci
ExecStop=-/usr/bin/nvidia-smi mig -dgi
ExecStop=/usr/bin/nvidia-smi -mig 0
TimeoutStartSec=300
TimeoutStopSec=300

[Install]
WantedBy=multi-user.target
`
