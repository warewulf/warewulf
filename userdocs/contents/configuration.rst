======================
Warewulf Configuration
======================

The default installation of Warewulf will put all of the configuration
files into ``/etc/warewulf/``. In that directory, you will find the
primary configuration files needed by Warewulf.

warewulf.conf
=============

The Warewulf configuration exists as follows in the current version of
Warewulf (4.5.0):

.. code-block:: yaml

   WW_INTERNAL: 45
   ipaddr: 10.0.0.1
   netmask: 255.255.252.0
   network: 10.0.0.0
   warewulf:
     port: 9873
     secure: false
     update interval: 60
     autobuild overlays: true
     host overlay: true
     syslog: false
   dhcp:
     enabled: true
     range start: 10.0.1.1
     range end: 10.0.1.255
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
   container mounts:
     - source: /etc/resolv.conf
       dest: /etc/resolv.conf
       readonly: true

Generally you can leave this file as is, as long as you set the
appropriate networking information. Specifically the following
configurations:

* ``ipaddr``: This is the control node's networking interface
  connecting to the cluster's **PRIVATE** network. This configuration
  must match the host's network IP address for the cluster's private
  interface.

* ``netmask``: Similar to the ``ipaddr``, this is the subnet mask for
  the cluster's **PRIVATE** network and it must also match the host's
  subnet mask for the cluster's private interface.

* ``dhcp:range start`` and ``dhcp:range end``: This address range must
  exist in the network defined above. If it is outside of this
  network, failures will occur. This specifies the range of addresses
  you want DHCP to use.

The other configuration options are usually not touched, but they are
explained as follows:

* ``*:enabled``: This can be used to disable Warewulf's control of a
  system service. This is useful if you want to manage that service
  directly.

* ``*:systemd name``: This is so Warewulf can control some of the
  host's services. For the distributions that we've built and tested
  this on, these will require no changes.

* ``warewulf:port``: This is the port that the Warewulf web server
  will be listening on. It is recommended not to change this so there
  is no misalignment with node's expectations of how to contact the
  Warewulf service.

* ``warewulf:secure``: When ``true``, this limits the Warewulf server
  to only respond to runtime overlay requests originating from a
  privileged port. This prevents non-root users from requesting the
  runtime overlay, which may contain sensitive information.

  When ``true``, ``wwclient`` uses TCP port 987.

  Changing this option requires rebuilding node overlays and rebooting
  compute nodes, to configure them to use a privileged port.

* ``warewulf:update interval``: This defines the frequency (in
  seconds) with which the Warewulf client on the compute node fetches
  overlay updates.

* ``warewulf:autobuild overlays``: This determines whether per-node
  overlays will automatically be rebuilt, e.g., when an underlying
  overlay is changed.

* ``warewulf:host overlay``: This determines whether the special
  ``host`` overlay is applied to the Warewulf server during
  configuration. (The host overlay is used to configure the dependent
  services.)

* ``warewulf:syslog``: This determines whether Warewulf server logs go
  to syslog or are written directly to a log file. (e.g.,
  ``/var/log/warewulfd.log``)

* ``nfs:export paths``: Warewulf can automatically set up these NFS
  exports.

* ``container mounts``: These paths are mounted into the container
  during ``container exec`` or ``container shell``, typically to allow
  them to operate in the host environment prior to deployment.

Paths
-----

*New in Warewulf v4.5.0*

Default paths to containers, overlays, and other Warewulf components
may be overridden using ``warewulf.conf:paths``.

.. code-block:: yaml

   paths:
     bindir: /usr/bin
     sysconfdir: /etc
     localstatedir: /var/lib
     ipxesource: /usr/share/ipxe
     srvdir: /var/lib
     firewallddir: /usr/lib/firewalld/services
     systemddir: /usr/lib/systemd/system
     wwoverlaydir: /var/lib/warewulf/overlays
     wwchrootdir: /var/lib/warewulf/chroots
     wwprovisiondir: /var/lib/warewulf/provision
     wwclientdir: /warewulf

nodes.conf
==========

The ``nodes.conf`` file is the primary database file for all compute
nodes. It is a flat text YAML configuration file that is managed by
the ``wwctl`` command, but some sites manage the compute nodes and
infrastructure via configuration management. This file being flat text
and very light weight makes management of the node configurations very
easy no matter what your configuration paradigm is.

For the purpose of this document, we will not go into the detailed
format of this file as it is recommended to edit with the ``wwctl``
command.

.. note::

   This configuration is not written at install time; but, the first
   time you attempt to run ``wwctl``, this file will be generated if
   it does not exist already.

defaults.conf
=============

The ``defaults.conf`` file configures default values used when none
are specified in ``nodes.conf``. For example: if a node does not have
a "runtime overlay" specified, the respective value from
``defaultnode`` is used. If a network device does not specify a
"device," the device value of the ``dummy`` device is used.

If ``defaults.conf`` does not exist, the following values are used as
compiled into Warewulf at build-time:

.. code-block:: yaml

  --
  defaultnode:
    runtime overlay:
    - generic
    system overlay:
    - wwinit
    kernel:
      args: quiet crashkernel=no vga=791 net.naming-scheme=v238
    init: /sbin/init
    root: initramfs
    ipxe template: default
    profiles:
    - default
    network devices:
      dummy:
        device: eth0
        type: ethernet
        netmask: 255.255.255.0

There should never be a need to change this file: all site-local
parameters should be specified using either nodes or profiles.

Directories
===========

The ``/etc/warewulf/ipxe/`` directory contains *text/templates* that
are used by the Warewulf configuration process to configure the
``ipxe`` service.
