===============
Disk Management
===============

Warewulf itself does not manage disks, partitions, or file systems directly, but provides structures in the configuration for these objects.
At the moment warewulf supports `ignition`_ to create the partitions and file systems.

.. _ignition: https://coreos.github.io/ignition/

.. note::

   It is not currently possible to manage the root file system with
   Warewulf.

Warewulf can be used, for example, to create ``swap`` partitions or ``/scratch`` file systems.

Requirements
============

Partition and file system creation requires both ``ignition`` and ``sgdisk`` to be installed in the container image.

Rocky Linux
-----------

.. code-block:: shell

   dnf install ignition gdisk

openSuse Leap
-------------

.. code-block:: shell

   zypper install ignition gptfdisk

Storage objects
===============

The format of the storage objects is inspired by ``butane/ignition``;
but, where ``butane/ignition`` uses lists for holding disks, partitions and file systems, Warewulf uses maps instead.

A node or profile can have several disks, where each disk is identified by the path to its block device.
Every disks holds a map to its partitions and a ``bool`` switch to indicate if an existing partition table should be overwritten if it does not matched the desired configuration.

Each partition is identified by its label.
The partition number can be omitted, but specifying it is recommended as ``ignition`` may fail without it.
Partition sizes should also be set (specified in MiB), except of the last partition:
if no size is given, the maximum available size is used.
Each partition has the switches ``should_exist`` and ``wipe_partition_entry`` which control the partition creation process. When omitting a partition number the `wipe_partition_entry` should be true, as this allows ignition to replace the existing partition.

File systems are identified by their underlying block device, preferably using the ``/dev/by-partlabel`` format.
Except for a ``swap`` partition, an absolute path for the mount point must be specified for each file system.
Depending on the container used, valid formats are ``btrfs``, ``ext3``, ``ext4``, and ``xfs``.
Each file system has the switch ``wipe_filesystem`` to control whether an existing file system is wiped.

Ignition Implementation
=======================

The ignition implementation uses systemd services, as the underlying ``sgdisk`` command relies on dbus notifications.
All necessary services are distributed by the ``ignition`` overlay and depends on the existence of the file ``/warewulf/ignition.json``.
This file is created by the template function ``{{ createIgnitionJson }}`` only if the configuration contains necessary specifications for disks, partitions, and file systems.
If the file ``/warewulf/ignition.json`` exists, the service ``ignition-disks-ww4.service`` calls the ignition binary which takes creates partitions and file systems.
A systemd ``.mount`` unit is created for each configured file system, which also creates the necessary mount points in the root file system.
These mount units are required by the enabled ``ww4-disks.target``.
Entries in ``/etc/fstab`` are created with the ``no_auto`` option so that file systems can be easily mounted.

Example disk configuration
==========================

The following command will create a ``/scratch`` file system on the node ``n01``.

.. code-block:: shell

   wwctl node set n01 \
     --diskname /dev/vda --diskwipe \
     --partname scratch --partcreate --partnumber 1 \
     --fsname scratch --fsformat btrfs --fspath /scratch

As this is a single file system, the partition number can be omitted.

A swap partition with 1Gig can be added with

.. code-block:: shell

   wwctl node set n01 \
     --diskname /dev/vda \
     --partname swap --partsize=1024 --partnumber 2 \
     --fsname swap --fsformat swap --fspath swap

which has the partition number ``1`` so that it will be added before the
``/scratch`` partition.

Wiping disks
============

Unless you specify the `--fswipe` flag for a filesystem, `ignition` will try to
reuse existing file systems. For empty disks this means that the desired configuration
is created and the filesystems are mounted; and so the `--fswipe` can be omitted so
data is on the disk isn't wiped.
If there are pre-existing partitions and filesystem on the disk, omitting the `--fswipe` may lead to the outcome that no filesystems are created and mounted.
In that case you should:
* wipe the existing data with the means of tools like `wipefs` or `dd` [#]_
* set the `--fswipe` flag and remove it after one reboot, if you want to keep
existing data on the disk.

.. [#] With `wipefs` you have to remove the filesystem *and* parition information. E.g. use `wipefs -fa /dev/vda*` to remove all filesystem information and partition information.

See also [ignition documentation](https://coreos.github.io/ignition/operator-notes/#filesystem-reuse-semantics) for additional information.

Troubleshooting
===============

If the partition creation didn't work as expected you have a few options to investigate:

- Add ``systemd.log_level=debug`` and or ``rd.debug`` to the kernelArgs of the node you're working on.
- After the next boot you should be able to find verbose information on the node with ``journalctl -u ignition-ww4-disks.service``.
- You could also check the content of ``/warewulf/ignition.json``.
- You could try to tinker with ``/warewulf/ignition.json`` calling

  .. code-block:: shell

     /usr/lib/dracut/modules.d/30ignition/ignition \
       --platform=metal \
       --stage=disks \
       --config-cache=/warewulf/ignition.json \
       --log-to-stdout

  after each iteration on the node directly until you find the settings you need.
  (Make sure to unmount all partitions if ``ignition`` was partially successful.)
- Sometimes you need to add ``should_exist: "true"`` for the swap partition as well.
