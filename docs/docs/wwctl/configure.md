---
id: configure
title: wwctl configure
---

Commands in this command group manage and initialize services Warewulf depends on, based on the configuration in `warewulf.conf`.

## dhcp
DHCP is a dependent service to Warewulf. This command will configure DHCP as defined.

### -s, --show
Show configuration (don't update)

### --persist
Persist the configuration and initialize the service

## hosts
Write out the /etc/hosts file based on the Warewulf template (hosts.tmpl) in the Warewulf configuration directory.

### -s, --show
Show configuration (don't update)

### --persist
Persist the configuration and initialize the service

## nfs
NFS is an optional dependent service of Warewulf, this tool will automatically configure NFS as per the configuration in the warewulf.conf file.

### -s, --show
Show configuration (don't update)

### --persist
Persist the configuration and initialize the service

## ssh
SSH is an optionally dependent service for Warewulf, this tool will automatically setup the ssh keys nodes using the 'default' system overlay as well as user owned keys.

### --persist
Persist the configuration and initialize the service

## tftp
TFTP is a dependent service of Warewulf, this tool will enable the tftp services on your Warewulf master.

### -s, --show
Show configuration (don't update)

### --persist
Persist the configuration and initialize the service
