---
id: container
title: wwctl container
---

Starting with version 4, Warewulf uses containers to build the bootable VNFS images for nodes to boot. These commands will help you import, management, and transform containers into bootable Warewulf VNFS images.

## build
This command will build a bootable VNFS image from an imported container image.

### -a, --all
(re)Build all VNFS images for all nodes

### -f, --force
Force rebuild, even if it isn't necessary

### --setdefault
Set this container for the default profile

## delete
This command will delete a container that has been imported into Warewulf.

## exec
This command will allow you to run any command inside of a given warewulf container. This is commonly used with an interactive shell such as ``/bin/bash`` to run a virtual environment within the container.

## import
This command will pull and import a container into Warewulf so it can be used as a source to create a bootable VNFS image.

### -f, --force
Force overwrite of an existing container

### -u, --update
Update and overwrite an existing container

### -b, --build
Build container when after pulling

### --setdefault
Set this container for the default profile

## list, ls
This command will show you the containers that are imported into Warewulf.
