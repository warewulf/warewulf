=====================================
EL9 Quickstart (Rocky Linux and RHEL)
=====================================

Install Warewulf and dependencies
=================================

.. code-block:: bash

   sudo dnf groupinstall "Development Tools"
   sudo dnf install epel-release
   sudo dnf config-manager --set-enabled crb
   sudo dnf install golang tftp-server dhcp-server nfs-utils gpgme-devel libassuan-devel

   git clone https://github.com/hpcng/warewulf.git
   cd warewulf
   make clean Defaults.mk \
    PREFIX=/usr \
    BINDIR=/usr/bin \
    SYSCONFDIR=/etc \
    DATADIR=/usr/share \
    LOCALSTATEDIR=/var/lib \
    SHAREDSTATEDIR=/var/lib \
    MANDIR=/usr/share/man \
    INFODIR=/usr/share/info \
    DOCDIR=/usr/share/doc \
    SRVDIR=/var/lib \
    TFTPDIR=/var/lib/tftpboot \
    SYSTEMDDIR=/usr/lib/systemd/system \
    BASHCOMPDIR=/etc/bash_completion.d/ \
    FIREWALLDDIR=/usr/lib/firewalld/services \
    WWCLIENTDIR=/warewulf
   make all
   sudo make install

Configure firewalld
===================

Restart firewalld to register the added service file, add the service
to the default zone, and reload.

.. code-block:: bash

   sudo systemctl restart firewalld
   sudo firewall-cmd --permanent --add-service warewulf
   sudo firewall-cmd --permanent --add-service nfs
   sudo firewall-cmd --permanent --add-service tftp
   sudo firewall-cmd --reload

Configure the controller
========================

Edit the file ``/etc/warewulf/warewulf.conf`` and ensure that you've
set the appropriate configuration parameters. Here are some of the
defaults for reference assuming that ``192.168.200.1`` is the IP
address of your cluster's private network interface:

.. code-block:: yaml

   ipaddr: 192.168.200.1
   netmask: 255.255.255.0
   warewulf:
     port: 9873
     secure: false
     update interval: 60
   dhcp:
     enabled: true
     range start: 192.168.200.10
     range end: 192.168.200.99
     template: default
     systemd name: dhcpd
   tftp:
     enabled: true
     tftproot: /var/lib/tftpboot
     systemd name: tftp
   nfs:
     systemd name: nfs-server
     exports:
       - /home
       - /var/warewulf

.. note::

   The DHCP range ends at ``192.168.200.99`` and as you will see
   below, the first node static IP address (post boot) is configured
   to ``192.168.200.100``.

Start and enable the Warewulf service
=====================================

.. code-block:: bash

   # Start and enable the warewulfd service
   sudo systemctl enable --now warewulfd

Configure system services automatically
=======================================

There are a number of services and configurations that Warewulf relies
on to operate.  If you wish to configure all services, you can do so
individually (omitting the ``--all``) will print a help and usage
instructions.

.. code-block:: bash

   sudo wwctl configure --all

.. note::

   If you just installed the system fresh and have SELinux enforcing,
   you may need to reboot the system at this stage to properly set the
   contexts of the TFTP contents. After rebooting, you might also need
   to run ``$ sudo restorecon -Rv /var/lib/tftpboot/`` if there are
   errors with TFTP still.

Pull and build the VNFS container (including the kernel)
========================================================

This will pull a basic VNFS container from Docker Hub and import the
default running kernel from the controller node and set both in the
"default" node profile.

.. code-block:: bash

   sudo wwctl container import docker://ghcr.io/hpcng/warewulf-rockylinux:9 rocky-9


Set up the default node profile
===============================

Node configurations can be set via node profiles. Each node by default
is configured to be part of the ``default`` node profile, so any
changes you make to that profile will affect all nodes.

The following command will set the container we just imported above to
the ``default`` node profile:

.. code-block:: bash

   sudo wwctl profile set --yes --container rocky-8 "default"

Next we set some default networking configurations for the first
ethernet device. On modern Linux distributions, the name of the device
is not critical, as it will be setup according to the HW
address. Because all nodes will share the netmask and gateway
configuration, we can set them in the default profile as follows:

.. code-block:: bash

   sudo wwctl profile set --yes --netdev eth0 --netmask 255.255.255.0 --gateway 192.168.200.1 "default"

Once those configurations have been set, you can view the changes by
listing the profiles as follows:

.. code-block:: bash

   sudo wwctl profile list -a

Add a node
==========

Adding nodes can be done while setting configurations in one
command. Here we are setting the IP address of ``eth0`` and setting
this node to be discoverable, which will then automatically have the
HW address added to the configuration as the node boots.

Node names must be unique. If you have node groups and/or multiple
clusters, designate them using dot notation.

Note that the full node configuration comes from both cascading
profiles and node configurations which always supersede profile
configurations.

.. code-block:: bash

   sudo wwctl node add n0000.cluster --ipaddr 192.168.200.100 --discoverable true

At this point you can view the basic configuration of this node by
typing the following:

.. code-block:: bash

   sudo wwctl node list -a n0000.cluster

Turn on your compute node and watch it boot!
