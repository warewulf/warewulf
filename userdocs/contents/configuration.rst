======================
Warewulf Configuration
======================

The default installation of Warewulf will put all of the configuration
files into ``/etc/warewulf/``. In that directory, you will find the
primary configuration files needed by Warewulf.

warewulf.conf
=============

The Warewulf configuration exists as follows in the current version of
Warewulf (4.3.0):

.. code-block:: yaml

   WW_INTERNAL: 43
   ipaddr: 192.168.200.1
   netmask: 255.255.255.0
   network: 192.168.200.0
   warewulf:
     port: 9873
     secure: false
     update interval: 60
     autobuild overlays: true
     host overlay: true
     syslog: false
   dhcp:
     enabled: true
     range start: 192.168.200.50
     range end: 192.168.200.99
     systemd name: dhcpd
   tftp:
     enabled: true
     systemd name: tftp
   nfs:
     enabled: true
     export paths:
     - path: /home
       export options: rw,sync
       mount options: defaults
       mount: true
     - path: /opt
       export options: ro,sync,no_root_squash
       mount options: defaults
       mount: false
     systemd name: nfs-server

Generally you can leave this file as is, as long as you set the
appropriate networking information. Specifically the following
configurations:

* ``ipaddr``: This is the control node's networking interface connecting
  to the cluster's **PRIVATE** network. This configuration must match
  the host's network IP address for the cluster's private interface.

* ``netmask``: Similar to the ``ipaddr``, this is the subnet mask for the
  cluster's **PRIVATE** network and it must also match the host's
  subnet mask for the cluster's private interface.

* ``dhcp:range start`` and ``dhcp:range end``: This address range must
  exist in the network defined above. If it is outside of this
  network, failures will occur. This specifies the range of addresses
  you want DHCP to use.

.. note::
   The network configuration listed above assumes the network
   layout in the [Background](background.md) portion of the
   documentation.

The other configuration options are usually not touched, but they are
explained as follows:

* ``*:enabled``: This disables Warewulf's control of an external
  service. This is useful if you want to manage that service directly.

* ``*:systemd name``: This is so Warewulf can control some of the host's
  services. For the distributions that we've built and tested this on,
  these will require no changes.

* ``warewulf:port``: This is the port that the Warewulf web server will
  be listening on. It is recommended not to change this so there is no
  misalignment with node's expectations of how to contact the Warewulf
  service.

* ``warewulf:secure``: When ``true``, this limits the Warewulf server to
  only respond to runtime overlay requests originating from a
  privileged report port. This makes it so that only the ``root`` user
  on a compute node can request the runtime overlay. While generally
  there is nothing super "secure" in these overlays, this adds the
  necessary protection that the user's can not obtain this
  information.

* ``warewulf:update interval``: This defines the frequency (in seconds)
  with which the Warewulf client on the compute node fetches overlay
  updates.
  
* ``warewulf:autobuild overlays``: This determines whether per-node
  overlays will automatically be rebuilt, e.g., when an underlying
  overlay is changed.
  
* ``warewulf:host overlay``: This determines whether the special ``host``
  overlay is applied to the Warewulf server during configuration. (The
  host overlay is used to configure the dependent services.)
  
* ``warewulf:syslog``: This determines whether Warewulf server logs go
  to syslog or are written directly to a log file. (e.g.,
  ``/var/log/warewulfd.log``)

* ``nfs:export paths``: Warewulf will automatically set up the NFS
  exports if you wish for it to do this.

nodes.conf
==========

The ``nodes.conf`` is the primary database file for all compute
nodes. It is a flat text YAML configuration file that is managed by
the ``wwctl`` command, but some sites manage the compute nodes and
infrastructure via configuration management. This file being flat text
and very light weight makes management of the node configurations very
easy no matter what your configuration paradigm is.

For the purpose of this document, we will not go into the detailed
format of this file as it is recommended to edit with the ``wwctl``
command.

.. note::
   This configuration is not written at install time, but the
   first time you attempt to run ``wwctl``, this file will be generated
   if it does not exist already.

Directories
===========

The ``/etc/warewulf/ipxe/`` contains *text/templates* that are used by
the Warewulf configuration process to configure the ``ipxe`` service.