#!/bin/csh

## Automatically configure SSH keys for a user on C SHell login
## Copy this file to /etc/profile.d along with ssh_setup.sh

set _UID=`id -u`

if ( $_UID < 500 && $_UID != 0 ) then
    exit
endif


if ( ! -f "$HOME/.ssh/config" && ! -f "$HOME/.ssh/cluster" ) then
    echo "Configuring SSH for cluster access"
    install -d -m 700 $HOME/.ssh
    ssh-keygen -t rsa -f $HOME/.ssh/cluster -N '' -C "Warewulf Cluster key" >& /dev/null
    cat $HOME/.ssh/cluster.pub >>! $HOME/.ssh/authorized_keys
    chmod 0600 $HOME/.ssh/authorized_keys

    touch $HOME/.ssh/config
    echo -n "# Added by Warewulf " >>! $HOME/.ssh/config
    (date +%Y-%m-%d >> $HOME/.ssh/config) |& /dev/null
    echo "Host *" >> $HOME/.ssh/config
    echo "   IdentityFile ~/.ssh/cluster" >> $HOME/.ssh/config
    echo "   StrictHostKeyChecking=no" >> $HOME/.ssh/config
    chmod 0600 $HOME/.ssh/config
endif
