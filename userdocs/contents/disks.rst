/Exa===============
Disk Management
===============

Warewulf itself does not manage disks, partitions, or file systems
directly, but provides structures in the configuration for these
objects. At the moment warewulf supports [ignition](https://coreos.github.io/ignition/) to create the partitions and file systems.

.. note::

   It is not currently possible to manage the root file system with
   Warewulf.

Warewulf can be used, for example, to create `swap` partitions or
`/scratch` file systems.

Requirements
===============

For the creation of partitions and file systems to work you will need to make sure that `ignition` as well as `sgdisk` is available in your container. 
 
 - `sgdisk` should be part of the package `gdisk` in RedHat-flavored OS like Rocky, AlmaLinux, etc. Therefore either add `dnf install gdisk` to your CI/CD-container-build-pipeline or manually install it in your container (`wwctl container shell <mycontainer>`).
 - check if `ignition` is present in the path `/usr/lib/dracut/modules.d/30ignition`. If it's missing you'll have shell into your container (`wwctl container shell <mycontainer>`) and run 
   ``` 
   git clone https://github.com/coreos/ignition.git
   dnf install go make libblkid-devel
   cd ignition 
   make 
   ``` 
 Once the build finished you'll find the binary in ./bin/. Proceed with
   ``` 
   mkdir -p /usr/lib/dracut/modules.d/30ignition
   cp bin/ignition /usr/lib/dracut/modules.d/30ignition/
   cd .. 
   rm -rf ignition 
   ```
 Now your container contains the module `ignition`. You may or may not want to remove `go`, `make`, and `libblkid-devel` before leaving the interactive container shell. 

Mind that the paths and/or package names may differ depending on the os you based your container on. 

Storage objects
===============

The format of the storage objects is inspired by `butane/ignition`;
but, where `butane/ignition` uses lists for holding disks, partitions
and file systems, Warewulf uses maps instead.

A node or profile can have several disks, where each disk is
identified by the path to its block device. Every disks holds a map to
its partitions and a `bool` switch to indicate if an existing
partition table should be overwritten if it does not matched the
desired configuration.

Each partition is identified by its label. The partition number can be
omitted, but specifying it is recommended as `ignition` may fail
without it. Partition sizes should also be set (specified in MiB),
except of the last partition: if no size is given, the maximum
available size is used. Each partition has the switches `should_exist`
and `wipe_partition_entry` which control the partition creation
process.

File systems are identified by their underlying block device,
preferably using the `/dev/by-partlabel` format. Except for a `swap`
partition, an absolute path for the mount point must be specified for
each file system. Depending on the container used, valid formats are
`btrfs`, `ext3`, `ext4`, and `xfs`. Each file system has the switch
`wipe_filesystem` to control whether an existing file system is wiped.

Ignition Implementation
=======================

The ignition implementation uses systemd services, as the underlying
`sgdisk` command relies on dbus notifications. All necessary services
are distributed by the `wwinit` overlay and depends on the existence
of the file `/warewulf/ignition.json`. This file is created by the
template function `{{ createIgnitionJson }}` only if the configuration
contains necessary specifications for disks, partitions, and file
systems.  If the file `/warewulf/ignition.json` exists, the service
`ignition-disks-ww4.service` calls the ignition binary which takes
creates partitions and file systems. A systemd `.mount` unit is
created for each configured file system, which also creates the
necessary mount points in the root file system. These mount units are
required by the enabled `ww4-disks.target`. Entries in `/etc/fstab`
are created with the `no_auto` option so that file systems can be
easily mounted.

Example disk configuration
==========================

The following command will create a `/scratch` file system on the node
`n01`

.. code-block:: shell

   wwctl node set n01 \
     --diskname /dev/vda --diskwipe \
     --partname scratch --partcreate \
     --fsname scratch --fsformat btrfs --fspath /scratch --fswipe

As this is a single file system, the partition number can be omitted.

A swap partition with 1Gig can be added with

.. code-block:: shell

   wwctl node set n01 \
     --diskname /dev/vda \
     --partname swap --partsize=1024 --partnumber 1 \
     --fsname swap --fsformat swap --fspath swap

which has the partition number `1` so that it will be added before the
`/scratch` partition.

Troubleshooting
===============

If the partition creation didn't work as expected you have a few options to investigate: 
 - add `systemd.log_level=debug` and or `rd.debug` to the kernelArgs of the node you're working on 
 - after the next boot you should be able to find verbose information on the node in the journal (`journalctl -u ignition-ww4-disks.service`). 
 - you could also check the content of `/warewulf/ignition.json`
 - you could try to tinker with `/warewulf/ignition.json` calling `/usr/lib/dracut/modules.d/30ignition/ignition --platform=metal --stage=disks --config-cache /warewulf/ignition.json -log-to-stdout` after each iteration on the node directly until you find the settings you need (make sure to unmount all partitions if `ignition` was partially successful). This would save you the time of the boot cycles. But you'll have to figure the analoge syntax in nodes.conf eventually.  
 - sometimes you need to add `should_exist: "true"` for the swap-partiton as well in `nodes.conf` either by calling `wwctl node edit` or by editing `nodes.conf` directly with your editor. 
