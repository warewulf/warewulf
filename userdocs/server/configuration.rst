====================
Server Configuration
====================

By default, the Warewulf server configuration is located at
``/etc/warewulf/warewulf.conf``. This is a YAML-formatted configuration file
used by to configured the Warewulf server itself and its external services.

An initial ``warewulf.conf`` is packaged with Warewulf. Each section is covered
in detail below.

Once Warewulf has been installed and configured:

* run ``wwctl configure --all`` to reconfigure external services
* run ``systemctl restart warewulfd`` to apply the configuration to the Warewulf
  server

Re-run both of these commands when making changes to ``warewulf.conf``.

.. code-block:: yaml

   ipaddr: 192.168.1.1
   netmask: 255.255.255.0
   network: 192.168.1.0
   warewulf:
     port: 9873
     secure: true
     update interval: 60
     autobuild overlays: true
     host overlay: true
     grubboot: false
   dhcp:
     enabled: true
     template: default
     systemd name: dhcpd
   tftp:
     enabled: true
     tftproot: /var/lib/tftpboot
     systemd name: tftp
     ipxe:
       "00:0B": arm64-efi/snponly.efi
       "00:00": undionly.kpxe
       "00:07": ipxe-snponly-x86_64.efi
       "00:09": ipxe-snponly-x86_64.efi
   nfs:
     enabled: true
     systemd name: nfsd
   ssh:
     key types:
       - ed25519
       - ecdsa
       - rsa
       - dsa
   image mounts:
     - source: /etc/resolv.conf
       dest: /etc/resolv.conf
   paths:
     bindir: /usr/bin
     sysconfdir: /etc
     localstatedir: /var/lib
     cachedir: /var/cache
     ipxesource: /usr/share/ipxe
     srvdir: /var/lib
     firewallddir: /usr/lib/firewalld/services
     systemddir: /usr/lib/systemd/system
     datadir: /usr/share
     wwoverlaydir: /var/lib/warewulf/overlays
     wwchrootdir: /var/lib/warewulf/chroots
     wwprovisiondir: /var/lib/warewulf/provision
     wwclientdir: /warewulf


warewulf
========

.. code-block:: yaml

   ipaddr: 192.168.1.1
   netmask: 255.255.255.0
   network: 192.168.1.0
   warewulf:
     port: 9873
     secure: true
     update interval: 60
     autobuild overlays: true
     host overlay: true
     grubboot: false

* ``ipaddr``: The Warewulf server address on the cluster network. This
  configuration must match the server's IP address.

  If ``ipaddr`` is specified as a CIDR address, ``netmask`` and ``network`` may
  be omitted.

* ``netmask``: The netmask for the cluster network.

* ``network``: The address of the cluster network itself.

* ``warewulf:port``: This is the port that the Warewulf web server will be
  listening on. It is recommended not to change this so there is no misalignment
  with node's expectations of how to contact the Warewulf service.

* ``warewulf:secure``: When ``true``, this limits the Warewulf server to only
  respond to runtime overlay requests originating from a privileged port. This
  prevents non-root users from requesting the runtime overlay, which may contain
  sensitive information.

  When ``true``, ``wwclient`` uses TCP port 987 by default. (A different port
  can be specified at ``wwclient:port``.)

  Changing this option requires rebuilding node overlays and rebooting compute
  nodes to configure them to use a privileged port for `wwclient`.

* ``warewulf:update interval``: This defines the frequency (in seconds) with
  which the Warewulf client on the compute node fetches overlay updates.

* ``warewulf:autobuild overlays``: Controls whether per-node overlays will
  automatically be rebuilt. (e.g., when an underlying overlay is changed)

  Overlay autobuild is not 100% reliable; but it is particularly useful for
  building overlays for new nodes.

* ``warewulf:host overlay``: Controls whether the special ``host`` overlay is
  applied to the Warewulf server during configuration. (The host overlay is used
  to configure external services.)

* ``warewulf::grubboot``: Controls whether iPXE (default) or GRUB is used as the
  network bootloader.

dhcp
====

The DHCP external service can be configured explicitly with ``wwctl configure
dhcp``. This (re)writes the DHCP configuration and enables and (re)starts the
DHCP service.

.. code-block:: yaml

   dhcp:
     enabled: true
     template: default
     systemd name: dhcpd

* ``dhcp:enabled``: Whether Warewulf should configure a DHCP server on the
  cluster network. Set to ``false`` when managing DHCP separately.

* ``dhcp:template`` An optional DHCP template variable to control the
  generation of the DHCP template.
  
  Specifying ``template: static`` populates ``dhcpd.conf`` with static leases
  for each host, bypassing the DHCP range. (Run ``wwctl configure dhcp`` to
  update ``dhcpd.conf`` when nodes are added, removed, or changed.)

* ``dhcp:range start`` and ``dhcp:range end``: Defines a dynamic DHCP range to
  use when provisioning cluster nodes. This address range must exist in the
  cluster network defined above. (Otherwise, the DHCP server will fail to
  start).

  This range should not overlap with IP addresses assigned to nodes in
  ``nodes.conf``.

* ``dhcp:systemd name``: Identifies the systemd service that manages the DHCP
  service. Used during ``wwctl configure dhcp`` to restart the service.

tftp
====

The TFTP external service can be configured explicitly with ``wwctl configure
tftp``. This writes the appropriate bootloader executables to the TFTP root
directory and enables the TFTP service.

.. code-block:: yaml

   tftp:
     enabled: true
     tftproot: /var/lib/tftpboot
     systemd name: tftp
     ipxe:
       "00:0B": arm64-efi/snponly.efi
       "00:00": undionly.kpxe
       "00:07": ipxe-snponly-x86_64.efi
       "00:09": ipxe-snponly-x86_64.efi

* ``tftp:enabled``: Whether Warewulf should configure a TFTP server on the
  cluster network. Set to ``false`` when managing TFTP separately.

* ``tftp:tftproot``: Identifies the local path being served by the managed TFTP
  server. Warewulf creates a ``warewulf/`` subdirectory and copies iPXE and/or
  GRUB bootloader files to this location depending on the server configuration.

* ``systemd name``: Identifies the systemd service that manages the TFTP
  service. Used during ``wwctl configure tftp`` to restart the service.

* ``ipxe``: A map of DHCP option architecture-types to the iPXE binary that
  should be used for that architecture. iPXE binaries are searched for in
  ``paths:ipxesource``. By default, these paths correspond to the location of
  the correct iPXE binary for each architecture in the distribution iPXE
  packages; but they can be specified explicitly when providing a local iPXE
  build.

nfs
===

The NFS external service can be configured explicitly with ``wwctl configure
nfs``. This configures the NFS server (particularly ``/etc/exports``) on the
Warewulf server and enables and starts the NFS service.

.. code-block:: yaml

   nfs:
     enabled: true
     export paths:
       - path: /home
         export options: rw,sync
       - path: /opt
         export options: ro,sync,no_root_squash
     systemd name: nfsd

* ``nfs:enabled``: Whether Warewulf should configure an NFS server on the
  cluster network. Set to ``false`` when not required or when managing NFS
  separately.

* ``nfs:export paths``: A list of NFS exports to configure on the Warewulf
  server. Each export defines a ``path`` to be exported and the ``export
  options`` for that export.

* ``systemd name``: Identifies the systemd service that manages the NFS
  service. Used during ``wwctl configure nfs`` to restart the service.

ssh
===

*New in Warewulf v4.5.1*

SSH key types to generate during ``wwctl configure ssh``. This create the
appropriate host keys (stored in ``/etc/warewulf/keys/``) and authentication
keys for passwordless ``ssh`` to cluster nodes. It also installs shell profiles
``/etc/profile.d/ssh_setup.csh`` and ``/etc/profile.d/ssh_setup.sh`` to
initialize authentication keys for new users if and when they log into the
Warewulf server.

.. code-block:: yaml

   ssh:
     key types:
       - ed25519
       - ecdsa
       - rsa
       - dsa

* ``ssh:key types``: Warewulf generate host keys for each listed key type.

The first listed key type is used to generate authentication ssh keys.

image mounts
============

A list of paths to temporarily mount from the Warewulf server into an image
during ``wwctl image exec`` and ``wwctl image shell``, typically to allow them
to operate in the host environment prior to deployment.

.. code-block:: yaml

   image mounts:
     - source: /etc/resolv.conf
       dest: /etc/resolv.conf

* ``image mounts:source``: The path on the Warewulf server to mount into the
  image.

* ``image mounts:dest``: The path in the image to use for the mount.

* ``image mounts::readonly``: Whether the mount should be read-only (``true``)
  or allow writes into the server path (``false``).

* ``image mounts::copy``: When ``true``, copy files into the image rather than
  mount. This is useful for initializing files with a starting value from the
  Warewulf server that should then be maintained as part of the image.

paths
=====

*New in Warewulf v4.5.0*

Override paths to images, overlays, and other Warewulf components.

.. code-block:: yaml

   paths:
     sysconfdir: /etc
     cachedir: /var/cache
     ipxesource: /usr/share/ipxe
     datadir: /usr/share
     wwoverlaydir: /var/lib/warewulf/overlays
     wwchrootdir: /var/lib/warewulf/chroots
     wwprovisiondir: /var/lib/warewulf/provision
     wwclientdir: /warewulf

* ``paths:sysconfdir``: The parent directory for the ``warewulf`` configuration
  directory, which stores ``warewulf.conf`` and ``nodes.conf``.

* ``paths::cachedir``: The parent directory for the ``warewulf`` cache of OCI
  images during ``wwctl image import``.

* ``paths:ipxesource``: Where to get iPXE binaries. These files are copied to
  ``warewulf.conf:tftp:tftproot`` by ``wwctl configure tftp``.

* ``datadir``: Parent directory for distribution overlays and BMC templates.

* ``paths:wwoverlaydir``: Parent directory for site overlays.

* ``paths:wwchrootdir``: Parent directory for Warewulf images.

* ``paths:wwprovisiondir``: The destination for built images and overlay images.

* ``paths:wwclientdir``: Where ``wwclient`` looks for its configuration on a
  provisioned node.

wwclient
========

Configuration for the ``wwclient`` service on cluster nodes.

.. code-block:: yaml

   wwclient:
     port: 987

* ``wwclient:port``: The source port used by ``wwclient``. By default an
  ephemeral port is selected; but ``warewulf.conf:warewulf:secure: true``
  requires a known privileged port.
  
  ``wwclient`` will use the TCP port "987" by default if ``secure: true``; but,
  if that port is otherwise in use, a different port may be specified.

hostfile
========

There are no explicit "hostfile" configuration options in ``warewulf.conf``; but
``wwctl configure hostfile`` updates the Warewulf server's ``/etc/hosts`` file
to include expected configuration for the server itself as well as the known
names of the cluster nodes and thier interfaces.

Entries from the Warewulf server's ``/etc/hosts`` file are distributed to
cluster nodes by the "hosts" overlay.
