package ssh_wwctl

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_sshWwctlOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.ImportFile("etc/warewulf/warewulf.conf", "warewulf.conf")
	assert.NoError(t, config.Get().Read(env.GetPath("etc/warewulf/warewulf.conf"), false))
	env.ImportFile("etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile("var/lib/warewulf/overlays/ssh.wwctl/rootfs/etc/profile.d/ssh_setup.sh.ww", "../rootfs/etc/profile.d/ssh_setup.sh.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.wwctl/rootfs/etc/profile.d/ssh_setup.csh.ww", "../rootfs/etc/profile.d/ssh_setup.csh.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "ssh.wwctl:ssh_setup.sh",
			args: []string{"--render", "host", "ssh.wwctl", "etc/profile.d/ssh_setup.sh.ww"},
			log:  ssh_setup_sh,
		},
		{
			name: "ssh.wwctl:ssh_setup.csh",
			args: []string{"--render", "host", "ssh.wwctl", "etc/profile.d/ssh_setup.csh.ww"},
			log:  ssh_setup_csh,
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
			assert.NoError(t, cmd.Execute())
			assert.Empty(t, stdout.String())
			assert.Empty(t, stderr.String())
			assert.Equal(t, tt.log, logbuf.String())
		})
	}
}

const ssh_setup_sh = `backupFile: true
writeFile: true
Filename: etc/profile.d/ssh_setup.sh
#!/bin/sh
##
## Copyright (c) 2001-2003 Gregory M. Kurtzer
##
## Copyright (c) 2003-2012, The Regents of the University of California,
## through Lawrence Berkeley National Laboratory (subject to receipt of any
## required approvals from the U.S. Dept. of Energy).  All rights reserved.
##
## Copied from https://github.com/warewulf/warewulf3/blob/master/cluster/bin/cluster-env

## Automatically configure SSH keys for a user on login
## Copy this file to /etc/profile.d

` + "_UID=`id -u`" + `
if [ $_UID -ge 500 -o $_UID -eq 0 ] && [ ! -f "$HOME/.ssh/config" -a ! -f "$HOME/.ssh/cluster" ]; then
    echo "Configuring SSH for cluster access"
    install -d -m 700 $HOME/.ssh
    ssh-keygen -t ed25519 -f $HOME/.ssh/cluster -N '' -C "Warewulf Cluster key" > /dev/null 2>&1
    cat $HOME/.ssh/cluster.pub >> $HOME/.ssh/authorized_keys
    chmod 0600 $HOME/.ssh/authorized_keys

    echo "# Added by Warewulf  ` + "`date +%Y-%m-%d 2>/dev/null`" + `" >> $HOME/.ssh/config
    echo "Host *" >> $HOME/.ssh/config
    echo "   IdentityFile ~/.ssh/cluster" >> $HOME/.ssh/config
    echo "   StrictHostKeyChecking=no" >> $HOME/.ssh/config
    chmod 0600 $HOME/.ssh/config
fi
`

const ssh_setup_csh = `backupFile: true
writeFile: true
Filename: etc/profile.d/ssh_setup.csh
#!/bin/csh

## Automatically configure SSH keys for a user on C SHell login
## Copy this file to /etc/profile.d along with ssh_setup.sh

` + "set _UID=`id -u`" + `
if ( ( $_UID > 500 || $_UID == 0 ) && ( ! -f "$HOME/.ssh/config" && ! -f "$HOME/.ssh/cluster" ) ) then
    echo "Configuring SSH for cluster access"
    install -d -m 700 $HOME/.ssh
    ssh-keygen -t ed25519 -f $HOME/.ssh/cluster -N '' -C "Warewulf Cluster key" >& /dev/null
    cat $HOME/.ssh/cluster.pub >>! $HOME/.ssh/authorized_keys
    chmod 0600 $HOME/.ssh/authorized_keys

    touch $HOME/.ssh/config
    echo -n "# Added by Warewulf " >>! $HOME/.ssh/config
    (date +%Y-%m-%d >> $HOME/.ssh/config) >& /dev/null
    echo "Host *" >> $HOME/.ssh/config
    echo "   IdentityFile ~/.ssh/cluster" >> $HOME/.ssh/config
    echo "   StrictHostKeyChecking=no" >> $HOME/.ssh/config
    chmod 0600 $HOME/.ssh/config
endif
`
