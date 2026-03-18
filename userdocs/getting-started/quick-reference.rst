===============
Quick Reference
===============

A quick reference for common day-to-day Warewulf operations.

Important paths
===============

.. code-block:: text

   # Configuration
   /etc/warewulf/warewulf.conf       # Main server configuration
   /etc/warewulf/nodes.conf          # Node database
   /etc/warewulf/auth.conf           # API authentication

   # Images (container filesystems)
   /var/lib/warewulf/chroots/        # Image root directories

   # Overlays
   /usr/share/warewulf/overlays/     # Distribution-provided overlays
   /var/lib/warewulf/overlays/       # Site-local overlays
   /var/lib/warewulf/provision/      # Built overlay images (per-node)

   # Logs
   journalctl -u warewulfd           # Warewulf daemon logs

Server management
=================

.. code-block:: shell

   # Configure all external services (TLS, warewulfd, TFTP, DHCP, NFS, SSH, hosts)
   wwctl configure --all

   # Configure individual services
   wwctl configure dhcp
   wwctl configure tftp
   wwctl configure nfs
   wwctl configure ssh
   wwctl configure hostfile

   # Restart the Warewulf daemon
   systemctl restart warewulfd

Listing nodes and profiles
===========================

.. code-block:: shell

   # List all nodes (summary)
   wwctl node list

   # List nodes with network interface details
   wwctl node list --net

   # List nodes with all fields (includes unset/inherited values)
   wwctl node list --all n1

   # List nodes with IPMI settings
   wwctl node list --ipmi

   # List specific nodes using hostlist syntax
   wwctl node list n[1-10]

   # Show current node boot status
   wwctl node status

   # List all profiles
   wwctl profile list

   # List all fields of a profile (includes inherited values)
   wwctl profile list default --all

Adding a node
=============

.. code-block:: shell

   # Add a single node with an IP address
   wwctl node add n1 --ipaddr=10.0.2.1

   # Add a range of nodes (IP address auto-increments)
   wwctl node add n[2-4] --ipaddr=10.0.2.2

   # Add a node with multiple options: image, network interface, kernel args, and profile
   wwctl node add n1 \
     --ipaddr=10.0.2.1 \
     --netmask=255.255.255.0 \
     --netdev=eno1 \
     --hwaddr=00:00:00:00:00:01 \
     --image=rockylinux-9 \
     --profile=default \
     --kernelargs="quiet crashkernel=no"

   # Un-set a field (revert to profile/default value)
   wwctl node set n1 --image=UNDEF

Network interfaces
==================

.. code-block:: shell

   # Set the primary network interface
   wwctl node set n1 \
     --netdev=eno1 \
     --hwaddr=00:00:00:00:00:01 \
     --ipaddr=10.0.2.1 \
     --netmask=255.255.255.0

   # Add a secondary network (e.g. InfiniBand)
   wwctl node set n1 \
     --netname=infiniband \
     --type=infiniband \
     --netdev=ib0 \
     --ipaddr=10.0.3.1 \
     --netmask=255.255.255.0

   # Configure a VLAN interface
   wwctl node set n1 \
     --netname=vlan42 \
     --netdev=vlan42 \
     --type=vlan \
     --ipaddr=10.0.42.1 \
     --netmask=255.255.252.0 \
     --nettagadd="vlan_id=42,parent_device=eth0"

   # Set DNS on an interface
   wwctl node set n1 --nettagadd="DNS1=1.1.1.1"

Images
======

.. code-block:: shell

   # Import an image from Docker Hub (or any OCI registry)
   wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:9 rockylinux-9

   # Import with registry credentials
   wwctl image import \
     --username myuser --password mysecret \
     docker://registry.example.com/myimage:latest myimage

   # Import from a local directory or tarball
   wwctl image import ./rockylinux-9/ rockylinux-9
   wwctl image import rockylinux-9.tar rockylinux-9

   # List available images
   wwctl image list

   # Open an interactive shell inside an image
   wwctl image shell rockylinux-9

   # Run a single command inside an image (e.g. install packages)
   wwctl image exec rockylinux-9 -- dnf -y install vim htop

   # Rebuild the image archive (after making changes)
   wwctl image build rockylinux-9

   # Assign an image to a node or profile
   wwctl node set n1 --image=rockylinux-9
   wwctl profile set default --image=rockylinux-9

IPMI
====

.. code-block:: shell

   # Configure IPMI settings on the default profile
   wwctl profile set default \
     --ipminetmask=255.255.255.0 \
     --ipmiuser=admin \
     --ipmipass=passw0rd \
     --ipmiinterface=lanplus \
     --ipmiwrite

   # Set the IPMI address on a specific node
   wwctl node set n1 --ipmiaddr=192.168.2.1

   # List IPMI settings
   wwctl node list --ipmi

   # Power commands
   wwctl power status n[1-10]
   wwctl power on n1
   wwctl power off n1
   wwctl power cycle n1
   wwctl power reset n1

   # Open serial-over-LAN console
   wwctl node console n1

Overlays
========

.. code-block:: shell

   # List all available overlays
   wwctl overlay list

   # Build overlays for all nodes
   wwctl overlay build

   # Build overlays for a specific node
   wwctl overlay build n1

   # Create a new site-local overlay
   wwctl overlay create myoverlay

   # Import a file from the host into an overlay
   wwctl overlay import myoverlay /etc/motd

   # Edit a file in an overlay (opens $EDITOR)
   wwctl overlay edit myoverlay /etc/motd

   # Show an overlay file (optionally rendered for a specific node)
   wwctl overlay show myoverlay /etc/motd
   wwctl overlay show myoverlay /etc/motd.ww --render=n1

   # Assign overlays to all nodes via the default profile
   wwctl profile set default \
     --system-overlays="wwinit,wwclient,fstab,hostname,ssh.host_keys,NetworkManager" \
     --runtime-overlays="hosts,ssh.authorized_keys"
