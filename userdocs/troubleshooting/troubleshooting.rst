===============
Troubleshooting
===============

sos
===

The ``warewulf-sos`` package (new in v4.6.1) adds support for gathering Warewulf
server configuration information in an sos report.

.. code-block::

   dnf -y install warewulf-sos
   sos report # optionally, --enable-plugins=warewulf

.. note::

   The ``warewulf-sos`` package is not currently built for SUSE.

warewulfd
=========

The Warewulf server (``warewulfd``) sends logs to the systemd journal.

.. code-block::

   journalctl -u warewulfd.service

To increase the verbosity of the log, specify either ``--verbose`` or
``--debug`` in the warewulfd OPTIONS.

.. code-block::

   echo "OPTIONS=--debug" >>/etc/default/warewulfd
   systemctl restart warewulfd.service

iPXE
====

If you're using iPXE to boot (the default), you can get a command prompt by
pressing with C-b during boot.

From the iPXE command prompt, you can run the same commands from `default.ipxe`_
to troubleshoot potential boot problems.

.. _default.ipxe: https://github.com/warewulf/warewulf/blob/main/etc/ipxe/default.ipxe

For example, the following commands perform a (relatively) normal Warewulf boot.
(Substitute your Warewulf server's IP address in place of 10.0.0.1, update the
port number if you have changed it from the default of 9873, and substitute your
cluster node's MAC address in place of 00:00:00:00:00:00.)

.. code-block::

   set uri http://10.0.0.1:9873/provision/00:00:00:00:00:00
   kernel --name kernel ${uri}?stage=kernel
   imgextract --name image ${uri}?stage=image&compress=gz
   imgextract --name system ${uri}?stage=system&compress=gz
   imgextract --name runtime ${uri}?stage=runtime&compress=gz
   boot kernel initrd=image initrd=system initrd=runtime

- The ``uri`` variable points to ``warewulfd`` for future reference. This
  includes the cluster node's MAC address so that Warewulf knows what image and
  overlays to provide.

- The ``kernel`` command fetches a kernel for later booting.

- The ``imgextract`` command fetches and decompresses the images that will make
  up the booted noe image. In a typical environment this is used to load a
  minimal "initial ramdisk" which, then, boots the rest of the system. Warewulf,
  by default, loads the entire image as an initial ramdisk, and also loads the
  system and runtime overlays at this time time.

- The ``boot`` command tells iPXE to boot the system with the given kernel and
  ramdisks.

.. note::

   This example does not provide ``assetkey`` information to ``warewulfd``. If
   your nodes have defined asset tags, provide it in the ``uri`` variable for
   the node you are trying to boot.

For example, you may want to try booting to a pre-init shell with debug logging
enabled. To do so, substitute the ``boot`` command above.

.. code-block::

   boot kernel initrd=image initrd=system initrd=runtime rdinit=/bin/sh

.. note::

   You may be more familiar with specifying ``init=`` on the kernel command
   line. ``rdinit`` indicates "ramdisk init." Since Warewulf, by default, boots
   the node image as an initial ramdisk, we must use ``rdinit=`` here.

GRUB
====

If you're using GRUB to boot, you can get a command prompt by pressing "c" when
prompted during boot.

From the GRUB command prompt, you can enter the same commands that you would
otherwise find in `grub.cfg.ww`_.

.. _grub.cfg.ww: https://github.com/warewulf/warewulf/blob/main/etc/grub/grub.cfg.ww

For example, the following commands perform a (relatively) normal Warewulf boot.
(Substitute your Warewulf server's IP address in place of 10.0.0.1, and update
the port number if you have changed it from the default of 9873.)

.. code-block::

   uri="(http,10.0.0.1:9873)/provision/${net_default_mac}"
   linux "${uri}?stage=kernel" wwid=${net_default_mac}
   initrd "${uri}?stage=image&compress=gz" "${uri}?stage=system&compress=gz" "${uri}?stage=runtime&compress=gz"
   boot

- The ``uri`` variable points to ``warewulfd`` for future reference.
  ``${net_default_mac}`` provides Warewulf with the MAC address of the booting
  node, so that Warewulf knows what image and overlays to provide it.

- The ``linux`` command tells GRUB what kernel to boot, as provided by
  ``warewulfd``. The ``wwid`` kernel argument helps ``wwclient`` identify the
  node during runtime.

- The ``initrd`` command tells GRUB what images to load into memory for boot. In
  a typical environment this is used to load a minimal "initial ramdisk" which,
  then, boots the rest of the system. Warewulf, by default, loads the entire
  image as an initial ramdisk, and also loads the system and runtime overlays at
  this time time.

- The ``boot`` command tells GRUB to boot the system with the previously-defined
  configuration.

.. note::

   This example does not provide ``assetkey`` information to ``warewulfd``. If
   your nodes have defined asset tags, provide it in the ``uri`` variable for
   the node you are trying to boot.

For example, you may want to try booting to a pre-init shell with debug logging
enabled. To do so, substitute the ``linux`` command above.

.. code-block::

   linux "${uri}?stage=kernel" wwid=${net_default_mac} debug rdinit=/bin/sh

.. note::

   You may be more familiar with specifying ``init=`` on the kernel command
   line. ``rdinit`` indicates "ramdisk init." Since Warewulf, by default, boots
   the node image as an initial ramdisk, we must use ``rdinit=`` here.

Dracut
======

By default, dracut simply panics and terminates when it encounters an issue.

Dracut looks at the kernel command line for its configuration. You can configure
it for additional logging and to switch to an interactive shell on error:

.. code-block::

   wwctl profile set default --kernelargs=rd.shell,rd.debug,log_buf_len=1M

For more information on debugging Dracut problems, see `the Fedora dracut
problems guide.`_

.. _the Fedora dracut problems guide.: https://docs.fedoraproject.org/en-US/quick-docs/debug-dracut-problems/

Ignition
========

If partition creation doesn't work as expected you have a few options to
investigate:

- Add ``systemd.log_level=debug`` and or ``rd.debug`` to the kernelArgs of the
  node you're working on.
- After the next boot you should be able to find verbose information on the node
  with ``journalctl -u ignition-ww4-disks.service``.
- You could also check the content of ``/warewulf/ignition.json``.
- You could try to tinker with ``/warewulf/ignition.json`` calling

  .. code-block:: shell

     /usr/lib/dracut/modules.d/30ignition/ignition \
       --platform=metal \
       --stage=disks \
       --config-cache=/warewulf/ignition.json \
       --log-to-stdout

  after each iteration on the node directly until you find the settings you
  need. (Make sure to unmount all partitions if ``ignition`` was partially
  successful.)
- Sometimes you need to add ``should_exist: "true"`` for the swap partition as
  well.

Running Containers on Cluster Nodes
===================================

Some container runtimes, notably Podman, require file system features that are
not available in ``initrootfs``. Cluster nodes using Podman (and some other
container runtimes) should be configured with ``--root=tmpfs``.

tmpfs and NUMA
==============

Warewulf can optionally mount the root filesystem as ``tmpfs`` instead of the
default ``initramfs``. Warewulf will add ``mpol=interleave`` to the mount point
which will distribute the memory across all NUMA nodes. This avoids the
hotspotting that occurs when the default initramfs stores large node images on a
single NUMA node. To enable this, set the rootfs type to tmpfs:

.. code-block:: shell
   
   wwctl profile set default --root=tmpfs
   
You may also adjust the tmpfs size via the ``wwinit.tmpfs.size`` kernel
argument:

.. code-block:: shell
   
   # Set tmpfs to use maximum 1GB  
   wwctl profile set default --kernelargs="wwinit.tmpfs.size=1G"
   # You can also use a percentage of physical RAM
   wwctl profile set default --kernelargs="wwinit.tmpfs.size=25%" 

By default this is set to 50% of physical RAM. Note that tmpfs is required for
SELinux overlays since initramfs cannot preserve SELinux contexts.

.. note::

   On some systems, it may also be necessary to include the ``noefi`` kernel
   argument. This works around specific EFI firmware bugs that can prevent
   proper memory release during the transition from ``initramfs`` to ``tmpfs``.
