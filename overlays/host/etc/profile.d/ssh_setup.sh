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

_UID=`id -u`

if [ $_UID -lt 500 -a $_UID -ne 0 ]; then
    exit
fi


if [ ! -f "$HOME/.ssh/config" -a ! -f "$HOME/.ssh/cluster" ]; then
    echo "Configuring SSH for cluster access"
    install -d -m 700 $HOME/.ssh
    ssh-keygen -t rsa -f $HOME/.ssh/cluster -N '' -C "Warewulf Cluster key" > /dev/null 2>&1
    cat $HOME/.ssh/cluster.pub >> $HOME/.ssh/authorized_keys
    chmod 0600 $HOME/.ssh/authorized_keys

    echo "# Added by Warewulf  `date +%Y-%m-%d 2>/dev/null`" >> $HOME/.ssh/config
    echo "Host *" >> $HOME/.ssh/config
    echo "   IdentityFile ~/.ssh/cluster" >> $HOME/.ssh/config
    echo "   StrictHostKeyChecking=no" >> $HOME/.ssh/config
    chmod 0600 $HOME/.ssh/config
fi
