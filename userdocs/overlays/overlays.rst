========
Overlays
========

Warewulf supplements provisioned node images with an "overlay" system. Overlays
are collections of files and :ref:`templates` that are rendered
and built per-node and then applied over the image during the provisioning
process.

Overlays are the primary mechanism for adding functionality Warewulf. Much of
even core functionality in Warewulf is implemented as distribution overlays, and
this flexibility is also available for local, custom overlays. By combining
templates with tags, network tags, and resources, the node registry
(``nodes.conf``) can become an expressive metadata store for arbitrary cluster
node configuration.

You can list the available overlays with ``wwctl overlay list``, and the files
within the overlays with ``wwctl overlay list --all``.

.. code-block:: console

   # wwctl overlay list --all fstab
   OVERLAY NAME  FILES/DIRS    SITE
   ------------  ----------    ----
   fstab         etc/          false
   fstab         etc/fstab.ww  false

Overlay Variables
-----------------

The command ``wwctl overlay info`` shows the variables used in an overlay
template, along with the help text for each variable.

.. code-block:: console

   # wwctl overlay info NetworkManager etc/NetworkManager/system-connections/ww4-managed.ww
   VARIABLE                        HELP                                           TYPE    OPTION
   --------                        ----                                           ----    ------
   $netdev.Device                  Set the device for given network               string  --netdev
   $netdev.Gateway                 Set the node's network device gateway          IP      --gateway
   $netdev.Hwaddr                  Set the device's HW address for given network  string  --hwaddr
   $netdev.Ipaddr                  IPv4 address in given network                  IP      --ipaddr
   $netdev.Ipaddr6                 IPv4 address in given network                  IP      --ipaddr
   $netdev.Ipaddr6                 IPv6 address                                   IP      --ipaddr6
   $netdev.MTU                     Set the mtu                                    string  --mtu
   $netdev.OnBoot.BoolDefaultTrue  Enable/disable network device (true/false)     WWbool  --onboot
   $netdev.Tags
   $netdev.Tags.DNSSEARCH
   $netdev.Tags.downdelay
   $netdev.Tags.master
   $netdev.Tags.miimon
   $netdev.Tags.mode
   $netdev.Tags.parent_device
   $netdev.Tags.updelay
   $netdev.Tags.vlan_id
   $netdev.Tags.xmit_hash_policy
   $netdev.Type                    Set device type of given network               string  --type

Structure
=========

An overlay is a directory that is applied to the root of a cluster node's
runtime file system. The overlay source directory should contain a single
``rootfs`` directory which represents the actual root directory for the overlay.

.. code-block:: none

  /usr/share/warewulf/overlays/issue
  └── rootfs
      └── etc
          └── issue.ww

Adding Overlays to Nodes
========================

A node or profile can configure an overlay in two different ways:

* An overlay can be configured to apply only during boot, along with the node
  image. These overlays are called **system overlays**.
* An overlay can be configured to also apply periodically while the system is
  running. These overlays are called **runtime overlays**.

.. code-block:: shell

   wwctl profile set default \
     --system-overlays="wwinit,wwclient,fstab,hostname,ssh.host_keys,systemd.netname,NetworkManager" \
     --runime-overlays="hosts,ssh.authorized_keys"

Multiple overlays can be applied to a single node, and overlays from multiple
profiles are appended together when applied to a single node.

Building Overlays
=================

Overlays are built (e.g., with ``wwctl overly build``) into compressed overlay
images for distribution to cluster nodes. These images typically match these two
use cases: system and runtime. As such, each cluster node typically has two
overlay images.

.. code-block:: console

   # wwctl overlay build
   Building system overlay image for n1
   Created image for n1 system overlay: /var/lib/warewulf/provision/overlays/n1/__SYSTEM__.img
   Compressed image for n1 system overlay: /var/lib/warewulf/provision/overlays/n1/__SYSTEM__.img.gz
   Building runtime overlay image for n1
   Created image for n1 runtime overlay: /var/lib/warewulf/provision/overlays/n1/__RUNTIME__.img
   Compressed image for n1 runtime overlay: /var/lib/warewulf/provision/overlays/n1/__RUNTIME__.img.gz

Overlay images for multiple node are built in parallel. By default, each CPU in
the Warewulf server will build overlays independently. The number of workers can
be specified with the ``--workers`` option.

Warewulf will attempt to build/update overlays as needed (configurable in the
``warewulf.conf``); but not all cases are detected, and manual overlay builds
are often necessary.

Creating and Modifying Overlays
===============================

You can add a new overlay to Warewulf with ``wwctl overlay create``.

.. code-block:: shell

   wwctl overlay create issue

A new overlay is just an empty directory. For it to be useful it needs to
contain some files.

For example, ``wwctl overlay import`` imports files from the Warewulf server
into the overlay.

.. code-block:: shell

   wwctl overlay import --parents issue /etc/issue

This imports ``/etc/issue`` from the Warewulf server into the new ``issue``
overlay.

.. note::

   The ``issue`` overlay already existed as a distribution overlay. Creating one
   shadows the distribution overlay with a new site overlay, allowing for local
   modification.

   Any modification to a distribution overlay first transparently creates a new
   site overlay and applies any changes there: distribution overlays should
   always remain unmodified.

You can also edit a new or existing overlay file in an interactive editor.

.. code-block:: shell

   wwctl overlay edit issue /etc/issue

Use ``wwctl overlay show`` to inspect the content of an overlay file.

.. code-block:: shell

   wwctl overlay show issue /etc/issue

Overlay files that end with ``.ww`` are templates. You can use ``wwctl overlay
show --render=<node>`` to show how a given template file would be rendered for
distribution to a given cluster node.

.. code-block:: shell

   wwctl overlay delete issue /etc/issue
   wwctl overlay import issue /etc/issue /etc/issue.ww
   wwctl overlay show issue /etc/issue.ww --render=n1

More information about templates is available in :ref:`its own section
<templates>`.

The content of the file for the given overlay is displayed with this command.
With the ``--render`` option a template is rendered as it will be rendered for
the given node. The node name is a mandatory argument to the ``--render`` flag.
Additional information for the file can be suppressed with the ``--quiet``
option.

.. note::

   It is not possible to delete files with an overlay.

Permissions
-----------

Overlay files are distributed to cluster nodes with the same user, group, and
mode that they have on the Warewulf server. Use ``wwctl overlay chown`` and
``wwctl overlay chmod`` to adjust them as necessary.

.. code-block:: shell

   wwctl overlay chown issue /etc/issue.ww root root
   wwctl overlay chmod issue /etc/issue.ww 0644


Distribution Overlays
=====================

Warewulf distinguishes between **distribution** overlays, which are included
with Warewulf, and **site** overlays, which are created or added locally. A site
overlay always takes precedence over a distribution overlay with the same name.
Any modification of a distribution overlay with ``wwctl`` actually makes changes
to an automatically-generated **site** overlay cloned from the distribution
overlay.

Site overlays are often stored at ``/var/lib/warewulf/overlays/``. Distribution
overlays are often stored at ``/usr/share/warewulf/overlays/``. But these paths
are dependent on compilation, distribution, packaging, and configuration
settings.

wwinit
------

The **wwinit** overlay performs initial configuration of the Warewulf node. Its
`wwinit` script runs before ``systemd`` or other init is called and contains all
configurations which are needed to boot.

In particular:

- Configure the loopback interface
- Configure the BMC based on the node's configuration
- Update PAM configuration to allow missing shadow entries
- Relabel the file system for SELinux

Other overlays can place scripts in one of two locations for additional pre-init
provisioning actions:

- **/warewulf/wwinit.d/:** executed in the initial root final system before the
  image is loaded into its final location. In a two-stage boot, these scripts
  are executed in the Dracut initramfs.

- **/warewulf/init.d/:** executed in the final root file system but before
  calling ``init``.

.. _wwclient:

wwclient
--------

All configured overlays are provisioned initially along with the node image
itself; but **wwclient** periodically fetches and applies the runtime overlay to
allow configuration of some settings without a reboot.

wwclient contacts the ``ipaddr`` value from ``warewulf.conf`` by default. This
can be overridden by specifying a ``WW_IPADDR`` environment variable, which can
be set via an overlay in ``/etc/default/wwclient``.

The default wwclient overlay contains a ``wwclient`` executable compiled for the
same architecture as the Warewulf server. Architecture-specific wwclient.aarch64
and wwclient.x86_64 overlays are available as well. This supports using wwclient
on cluster nodes with a different architecture than the Warewulf server.

Network interfaces
------------------

Warewulf ships with support for many different network interface configuration
systems. All of these are applied by default; but the list may be trimmed to
the desired system.

- ifcfg
- ifupdown
- NetworkManager
- wicked

.. note::

   The ``ifupdown`` overlay was previously named ``debian.interfaces``. The old
   name is still supported for compatibility, but it is deprecated and will be
   removed in a future release.

Warewulf also configures both systemd and udev with the intended names of
configured network interfaces, typically based on a known MAC address.

- systemd.netname
- udev.netname

.. _dns:

Several of the network configuration overlays support netdev tags to further
customize the interface:

- **DNS[0-9]*:** one or more DNS servers
- **DNSSEARCH:** domain search path
- **MASTER:** the master for a bond interface

NetworkManager
^^^^^^^^^^^^^^

- **parent_device:** the parent device of a vlan interface
- **vlan_id:** the vlan id for a vlan interface
- **downdelay, updelay, miimon, mode, xmit_hash_policy:**
  bond device settings

Basics
------

The **hostname** overlay sets the hostname based on the configured Warewulf
node name.

The **hosts** overlay configures ``/etc/hosts`` to include all Warewulf nodes.

The **issue** overlay configures a standard Warewulf status message for display
during login.

The **resolv** overlay configures ``/etc/resolv.conf`` based on the value of
"DNS" :ref:`nettags <nettags>`. (In most situations this should be unnecessary,
as the network interface configuration should handle this dynamically.)

.. code-block:: shell

   wwctl node set n1 --nettagadd="DNS1=1.1.1.1"
   wwctl node set n1 --nettagadd="DNS2=1.0.0.1"

fstab
-----

The **fstab** overlay configures ``/etc/fstab`` based on the data provided in the "fstab"
resource. It also creates entries for file systems defined by Ignition.

.. code-block:: yaml

   nodeprofiles:
     default:
       resources:
         fstab:
           - spec: warewulf:/home
             file: /home
             vfstype: nfs
           - spec: warewulf:/opt
             file: /opt
             vfstype: nfs

ssh
---

Two SSH overlays configure host keys (one set for all node in the cluster) and
``authorized_keys`` for the root account.

- ssh.authorized_keys
- ssh.host_keys

.. _Syncuser:

syncuser
--------

The **syncuser** overlay updates ``/etc/passwd`` and ``/etc/group`` to include
all users on both the Warewulf server and from the image.

To function properly, ``wwctl image syncuser`` (or the ``--syncuser`` option
during ``import``, ``exec``, ``shell``, or ``build``) must have also been run on
the image to synchronize its user and group IDs with those of the server.

If a ``PasswordlessRoot`` tag is set to "true", the overlay will also insert a
"passwordless" root entry. This can be particularly useful for accessing a
cluster node when its network interface is not properly configured.

.. warning::

   ``PasswordlessRoot`` is not recommended for production; it should only be
   used during debugging, when normal authentication is not functional.

ignition
--------

The **ignition** overlay defines partitions and file systems on local disks.
Configuration may be provided via native disk, partition, and filesystem fields
or via an ``ignition`` resource.

.. code-block:: yaml

   ignition:
     storage:
       disks:
         - device: /dev/vda
           partitions:
             - label: scratch
               shouldExist: true
               wipePartitionEntry: true
           wipeTable: true
       filesystems:
         - device: /dev/disk/by-partlabel/scratch
           format: btrfs
           path: /scratch
           wipeFilesystem: false

If any disk/partition/filesystem configuration is provided for a node with
explicit arguments to ``wwctl <node|profile> set``, the ``ignition`` resource is
ignored.

To use ignition during Dracut (so that the root file system may be provisioned
before the image is loaded) include Ignition in the Dracut image.

.. code-block:: shell

   wwctl image exec rockylinux-9 -- /usr/bin/dracut --force --no-hostonly --add wwinit --add ignition --regenerate-all

debug
-----

The **debug** overlay is not intended to be used in configuration, but is
provided as an example. In particular, the provided `tstruct.md.ww` demonstrates
the use of most available template metadata.

.. code-block:: shell

   wwctl overlay show --render=<nodename> debug tstruct.md.ww

.. _localtime:

localtime
---------

The **localtime** overlay configures the timezone of a cluster node to match
that of the Warewulf server; alternatively, a different timezone may be
specified with a ``localtime`` tag.

.. code-block:: shell

   wwctl profile set default --tagadd="localtime=UTC"

sfdisk
------

The **sfdisk** overlay partitions block devices during wwinit. Configuration may
be provided using native disk and partition configuration or via an ``sfdisk``
resource.

Multiple devices can be partitioned, with each device provided as an item in
``sfdisk`` list.

.. code-block:: yaml

   sfdisk:
     - device: /dev/sda
       label: gpt
       partitions:
         - device: /dev/sda1
           name: sfdisk-rootfs
           size: 4194304
         - device: /dev/sda2
           name: sfdisk-scratch
           size: 1048576
         - device: /dev/sda3
           name: sfdisk-swap
           size: 2097152

All headers and named partition fields supported by the ``sfdisk`` input format
are supported in the ``sfdisk`` resource.

If any disk/partition configuration is provided for a node with explicit
arguments to ``wwctl <node|profile> set``, the ``sfdisk`` resource is ignored.

To use the sfdisk overlay, include sfdisk in the Dracut image. Optionally also
include blockdev and/or udevadm, to allow the partition table to be re-scanned.

.. code-block:: shell

   wwctl image exec rockylinux-8 -- /usr/bin/dracut --force --no-hostonly --add wwinit --install sfdisk --install blockdev --install udevadm --regenerate-all

For more information, see the :ref:`provision to disk` section.

mkfs
----

The **mkfs** overlay formats block devices during wwinit. Configuration may be
provided using native filesystem fields or via an ``mkfs`` resource.

.. code-block:: yaml

   mkfs:
     - device: /dev/sda1 # the device to format
       type: xfs # what type of file system to create
       options: -b 1024 # additional options to pass to mkfs
       overwrite: false # whether to overwrite an existing format
       size: 0 # defaults to the full device

If any filesystem configuration is provided for a node with explicit arguments
to ``wwctl <node|profile> set``, the ``mkfs`` resource is ignored.

To use the mkfs overlay, include mkfs and any necessary file-system-specific
sub-commands in the Dracut image. Optionally also include wipefs to detect
existing file systems.

.. code-block:: shell

   wwctl image exec rockylinux-9 -- /usr/bin/dracut --force --no-hostonly --add wwinit --install mkfs --install mkfs.ext4 --install wipefs --regenerate-all

For more information, see the :ref:`provision to disk` section.

mkswap
------

The **mkswap** overlay formats block devices during wwinit. Configuration may be
provided using native filesystem fields or via a ``mkswap`` resource.

.. code-block:: yaml

   mkswap:
     - device: /dev/sda2 # the device to format
       overwrite: false # whether to overwrite an existing format
       label: swap # the label to set for the swap device

If any filesystem configuration is provided for a node with explicit arguments
to ``wwctl <node|profile> set``, the ``mkswap`` resource is ignored.

To use the mkswap overlay, include mkswap in the Dracut image. Optionally also
include wipefs to detect existing file systems.

.. code-block:: shell

   wwctl image exec rockylinux-9 -- /usr/bin/dracut --force --no-hostonly --add wwinit --install mkswap --regenerate-all

systemd mounts
--------------

Two overlays, **systemd.mount** and **systemd.swap**, configure mounted and swap
storage based on the configuration of native file system fields. They are often
paired with the ``mkfs`` and ``mkswap`` overlays.

host
----

Configuration files used for the configuration of the Warewulf host /
server are stored in the **host** overlay. Unlike other overlays, it
*must* have the name ``host``. Existing files on the host are copied
to backup files with a ``wwbackup`` suffix at the first
run. (Subsequent use of the host overlay won't overwrite existing
``wwbackup`` files.)

The following services get configuration files via the host overlay:

* ssh keys are created with the scrips ``ssh_setup.sh`` and
  ``ssh_setup.csh``
* hosts entries are created by manipulating ``/etc/hosts`` with the
  template ``hosts.ww``
* nfs kernel server receives its exports from the template
  ``exports.ww``
* the dhcpd service is configured with ``dhcpd.conf.ww``
