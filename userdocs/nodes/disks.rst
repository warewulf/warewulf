==================
Provisioning disks
==================

As a tech preview, Warewulf provides structures to define disks, partitions, and
file systems. These structures can generate a configuration for `Ignition`_ to
provision partitions and file systems dynamically on cluster nodes, or with
sfdisk, mkfs, and mkswap during boot.

.. _Ignition: https://coreos.github.io/ignition/

Ignition can, for example, create ``swap`` partitions or ``/scratch`` file
systems.

.. note::

   Warewulf is not currently able to provision the node image onto an explicitly
   provisioned root file system.

Requirements
============

Partition and file system creation requires both ``ignition`` and ``sgdisk`` to
be installed in the image.

Rocky Linux
-----------

.. code-block:: shell

   dnf install ignition gdisk

.. note::

   Packages for Ignition are not currently available for Rocky Linux 8, but it
   is available for Rocky Linux 9 as part of "appstream."

openSuse Leap
-------------

.. code-block:: shell

   zypper install ignition gptfdisk

Disks and partitions
====================

A node or profile can have several disks. Each disk is identified by the path to
its block device. Each disk holds a map to its partitions and a ``bool`` switch
to indicate if an existing, non-matching partition table should be overwritten.

Each partition is identified by its label. The partition number can be omitted,
but specifying it is recommended as Ignition may fail without it. Partition
sizes should also be set (specified in MiB), except for the last partition: if
no size is given, the maximum available size is used. Each partition has the
switches ``should_exist`` and ``wipe_partition_entry`` which control the
partition creation process. When omitting a partition number the
`wipe_partition_entry` should be true, as this allows ignition to replace the
existing partition.

.. code-block:: shell

   wwctl node set n1 \
     --diskname /dev/vda --diskwipe \
     --partname scratch --partcreate --partnumber 1

File systems
============

File systems are identified by their underlying block device, preferably using
the ``/dev/by-partlabel`` format. Except for a ``swap`` partition, an absolute
path for the mount point must be specified for each file system. Depending on
the image used, valid formats are ``btrfs``, ``ext3``, ``ext4``, and ``xfs``.
Each file system has the switch ``wipe_filesystem`` to control whether an
existing file system is wiped.

.. code-block:: shell

   wwctl node set n1 \
     --diskname /dev/vda --partname scratch \
     --fsname scratch --fsformat btrfs --fspath /scratch

Boot-time configuration
=======================

Ignition uses systemd, as the underlying ``sgdisk`` command relies on dbus
notifications.


1. ``ignition-disks-ww4.service`` uses Ignition to create the specified
   partitions and file systems.

2. ``ww4-disks.target`` depends on a matching ``.mount`` unit for each
   mounted file system.
   
3. Each ``.mount`` creates the necessary mount points in the root file system
   and mounts the provisioned file systems during boot.

These services and mount units are generated by the ``ignition`` overlay and
depend on the existence of the file ``/warewulf/ignition.json``, also generated
by the ``ignition`` overlay.

Example disk configurations
===========================

This command formats a btrfs file system on a "scratch" partion of
"vda" and mounts it at ``/scratch``.

.. code-block:: shell

   wwctl node set n1 \
     --diskname /dev/vda --diskwipe \
     --partname scratch --partcreate --partnumber 1 \
     --fsname scratch --fsformat btrfs --fspath /scratch

This command adds a swap partition to the "vda" disk.

.. code-block:: shell

   wwctl node set n1 \
     --diskname /dev/vda \
     --partname swap --partsize=1024 --partnumber 2 \
     --fsname swap --fsformat swap --fspath swap

Re-using or wiping disks
========================

For empty disks the desired configuration is created and the filesystems are
mounted. If partitions or file systems already exist on the disk, ``ignition``
tries to reuse existing file systems by default.

To ignore existing file systems and provision fresh file systems on each boot,
specify the ``--fswipe``` flag for that filesystem, and ``--diskwipe`` for the
disk, as necessary.

If you would like to re-use existing partitions but want to replace existing
file systems once, you may

* wipe the existing data with tools like ``wipefs`` or `dd` [#]_; or
* set the ``--fswipe`` flag and remove it after one reboot.

.. [#] With ``wipefs`` you have to remove the filesystem *and* parition
    information. E.g., use ``wipefs -fa /dev/vda*`` to remove all file system
    information and partition information.

See the `upstream ignition documentation`_ for additional information.

.. _upstream ignition documentation: https://coreos.github.io/ignition/operator-notes/#filesystem-reuse-semantics


.. _provision to disk:

Provision to disk
=================

*New in Warewulf v4.6.2*

As a tech preview, the Warewulf two-stage boot process can provision the node
image to local storage.

.. warning::

   This functionality is a technology preview and should be used with care. Pay
   specific attention to ``wipeFilesytem`` and similar settings.

.. note::

   Warewulf doesn't install a bootloader to the disk or add UEFI entries. Nodes
   still request an image and configuration from the Warewulf server on every
   boot.

.. note::

   While provisioning to disk should be possible during a single-stage boot, not
   all features are available:

   - Warewulf does not perform hardware detection to ensure that necessary
     kernel modules are loaded prior to init.
   - Warewulf does not load udev to ensure that ``/dev/disk/by-*`` symlinks are
     available prior to init.

With Ignition
-------------

Warewulf needs a prepared file system to deploy the image to. Warewulf can
provision this file system using Ignition. To use Ignition, include ``ignition``
in your system overlay. The ignition overlay provisions disks during init and,
optionally, during the first stage of a two-stage boot. This allows the
root file system to be provisioned before the image is loaded.

.. code-block:: shell

   wwctl node set wwnode1 \
     --diskname /dev/vda --diskwipe \
     --partname rootfs --partcreate --partnumber 1 \
     --fsname rootfs --fsformat ext4 --fspath /

In order to allow Dracut to provision the disk, partition, and file system,
Ignition must be included in the Dracut image.

.. code-block:: shell

   wwctl image exec rockylinux-9 -- /usr/bin/dracut --force --no-hostonly --add wwinit --add ignition --regenerate-all

The necessary file system may alternatively be prepared out-of-band.

With sfdisk and mkfs
--------------------

Systems that do not have access to Ignition (e.g., Rocky Linux 8) can provision
the root file system using a combination of ``sfdisk`` and ``mkfs``. To use
them, include ``sfdisk`` and ``mkfs`` in your system overlay. The ``sfdisk`` and
``mkfs`` overlays provision disk and file systems during the first stage of a
two-stage boot. This allows the root file system to be provisioned before the
image is loaded.

Configure the ``sfdisk`` and ``mkfs`` overlays using resources:

.. code-block:: shell

   wwctl node set wwnode1 \
     --diskname /dev/vda --diskwipe \
     --partname rootfs --partcreate --partnumber 1 \
     --fsname rootfs --fsformat ext4 --fspath /

In order to allow Dracut to provision the disk, partition, and file system, some
additional commands must be included in the Dracut image, depending on which
functionality is used:

- **sfdisk:** writes the partition table

  - **blockdev:** used to re-read the partition table after writing

  - **udevadm:** used to trigger udev events after writing the partition table

- **mkfs:** formats file systems (may also require file-system-specific commands like mkfs.ext4)

  - **mkfs.ext4**, **mkfs.btrfs**, etc: used by mkfs to format specific file systems

  - **wipefs:** used to determine if a file system already exists

.. code-block:: shell

   wwctl image exec rockylinux-8 -- /usr/bin/dracut --force --no-hostonly \
     --add wwinit \
     --install sfdisk \
     --install blockdev \
     --install udevadm \
     --install mkfs \
     --install mkfs.ext4 \
     --install wipefs \
     --regenerate-all

Configuring the root device
---------------------------

Set the desired storage device for the node image using the ``--root``
parameter.

.. code-block:: shell

   wwctl node set wwnode1 --root /dev/disk/by-partlabel/rootfs
