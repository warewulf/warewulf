---
id: profile
title: wwctl profile
---

Management of node profile settings

## add
This command will add a new node profile.

## delete
This command will delete a node profile.

## list
This command will list and show the profile configurations.

## set
This command will allow you to set configuration properties for node profiles.

### --comment
Set a comment for this node

### -C, --container
Set the container (VNFS) for this node

### -K, --kernel
Set Kernel version for nodes

### -A, --kernelargs
Set Kernel argument for nodes

### -c, --cluster
Set the node's cluster group

### -P, --ipxe
Set the node's iPXE template name

### -i, --init
Define the init process to boot the container

### --root
Define the rootfs

### -R, --runtime
Set the node's runtime overlay

### -S, --system
Set the node's system overlay

### --ipminetmask
Set the node's IPMI netmask

### --ipmigateway
Set the node's IPMI gateway

### --ipmiuser
Set the node's IPMI username

### --ipmipass
Set the node's IPMI password

### -N, --netdev
Define the network device to configure

### -I, --ipaddr
Set the node's network device IP address

### -M, --netmask
Set the node's network device netmask

### -G, --gateway
Set the node's network device gateway

### -H, --hwaddr
Set the node's network device HW address

### --netdel
Delete the node's network device

### --netdefault
Set this network to be default

### -a, --all
Set all profiles

### -f, --force
Force configuration (even on error)
