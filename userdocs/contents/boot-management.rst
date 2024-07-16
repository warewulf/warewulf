===============
Boot Management
===============

Warewulf uses iPXE to for network boot by default. As a tech preview, support
for GRUB is also available, which adds support for secure boot.

Also as a tech preview, Warewulf may also use iPXE to boot a dracut
initramfs as an initial stage before loading the container image.

Booting with iPXE
=================

.. graphviz::

  digraph G{
      node [shape=box];
      compound=true;
      edge [label2node=true]
      bios [shape=record label="{BIOS | boots from DHCP/next-server via TFTP}"]

      subgraph cluster0 {
       label="iPXE boot"
       iPXE;
       ipxe_cfg [shape=record label="{ipxe.cfg|generated for each node}"];
       iPXE -> ipxe_cfg [label="http"];
      }

      bios->iPXE [lhead=cluster0,label="iPXE.efi"];

      kernel [shape=record label="{kernel|ramdisk (root fs)|wwinit overlay}|extracted from node container"];
      ipxe_cfg->kernel[ltail=cluster0,label="http"];
  }

Booting with GRUB
=================

Support for GRUB as a network bootloader (replacing iPXE) is available in
Warewulf as a technology preview.

.. graphviz::

  digraph G{
      node [shape=box];
      compound=true;
      edge [label2node=true]
      bios [shape=record label="{BIOS | boots from DHCP/next-server via TFTP}"]

      bios->shim [lhead=cluster1,label="shim.efi"];
      subgraph cluster1{
        label="Grub boot"
        shim[shape=record label="{shim.efi|from ww4 host}"];
        grub[shape=record label="{grubx64.efi | name hardcoded in shim.efi|from ww4 host}"]
        shim->grub[label="TFTP"];
        grubcfg[shape=record label="{grub.cfg|static under TFTP root}"];
        grub->grubcfg[label="TFTP"];
      }
      kernel [shape=record label="{kernel|ramdisk (root fs)|wwinit overlay}|extracted from node container"];
      grubcfg->kernel[ltail=cluster1,label="http"];
  }

Instead of the iPXE starter a combination of `shim and GRUB
<https://www.suse.com/c/uefi-secure-boot-details/>`_ can be used with the
advantage that secure boot can be used. That means that only the signed kernel
of a distribution can be booted. This can be a huge security benefit for some
scenarios.

In order to enable the grub boot method it has to be enabled in `warewulf.conf`.

.. code-block:: yaml

   warewulf:
     grubboot: true

Nodes which are not known to Warewulf are booted with the shim/grub from the
Warewulf server host.

Secure boot
-----------

.. graphviz::

   digraph foo {
      node [shape=box];
      subgraph boot {
        "EFI" [label="EFI",row=boot];
        "Shim" [label="Shim",row=boot];
        "Grub" [label="Grub",row=boot];
        "Kernel" [label="kernel",row=boot];
        EFI -> Shim[label="Check for Microsoft signature"];
        Shim -> Grub[label="Check for Distribution signature"];
        Grub->Kernel[label="Check for Distribution or MOK signature"];
      }
    }

If secure boot is enabled at every step a signature is checked and the boot
process fails if this check fails. The shim typically only includes the key for
a single operating system, which means that each distribution needs separate
`shim` and `grub` executables. Warewulf extracts these binaries from the
containers. If the node is unknown to Warewulf or can't be identified during
the TFTP boot phase, the shim/grub binaries of the host in which Warewulf is
running are used.

Install shim and efi
--------------------

`shim.efi` and `grub.efi` must be installed in the container for it to be
booted by GRUB.

.. code-block:: console

  # wwctl container shell leap15.5
  [leap15.5] Warewulf> zypper install grub2 shim

  # wwctl container shell rocky9
  [rocky9] Warewulf> dnf install shim-x64.x86_64 grub2-efi-x64.x86_64

These packages must also be installed on the Warewulf server host to enable
node discovery using GRUB.

http boot
---------

Modern EFI systems have the possibility to directly boot per http. The flow diagram
is the following:

.. graphviz::

  digraph G{
      node [shape=box];
      efi [shape=record label="{EFI|boots from URI defined in filename}"];
      shim [shape=record label="{shim.efi|replaces shim.efi with grubx64.efi in URI|extracted from node container}"];
      grub [shape=record label="{grub.efi|checks for grub.cfg|extracted from node container}"]
      kernel [shape=record label="{kernel|ramdisk (root fs)|wwinit overlay}|extracted from node container"];
      efi->shim [label="http"];
      shim->grub [label="http"];
      grub->kernel [label="http"];
    }

Warewulf delivers the initial `shim.efi` and `grub.efi` via http as taken
directly from the node's assigned container.

Booting with dracut
===================

Some systems, typically due to limitations in their BIOS or EFI
firmware, are unable to load container image of a certain size
directly with a traditional bootloader, either iPXE or GRUB. As a
workaround for such systems, Warewulf can be configured to load a
dracut initramfs from the container and to use that initramfs to load
the full container image.

Warewulf provides a dracut module to configure the dracut initramfs to
load the container image. This module is available in the
``warewulf-dracut`` subpackage, which must be installed in the
container image.

With the ``warewulf-dracut`` package installed, you can build an
initramfs inside the container.

.. code-block:: shell

   dnf -y install warewulf-dracut
   dracut --force --no-hostonly --add wwinit --kver $(ls /lib/modules | head -n1)

Set the node's iPXE template to ``dracut`` to direct iPXE to fetch the
node's initramfs image and boot with dracut semantics, rather than
booting the node image directly.

.. note::

   Warewulf iPXE templates are located at ``/etc/warewulf/ipxe/`` when
   Warewulf is installed via official packages. You can learn more
   about how dracut booting works by inspecting its iPXE template at
   ``/etc/warewulf/ipxe/dracut.ipxe``.

.. code-block:: shell

   wwctl node set wwnode1 --ipxe dracut

.. note::

   The iPXE template may be set at the node or profile level.

Alternatively, to direct GRUB to fetch the node's initramfs image and boot with
dracut semantics, set a ``GrubMenuEntry`` tag for the node.

.. note::

   Warewulf configures GRUB with a template located at
   ``/etc/warewulf/grub/grub.cfg.ww``. Inspect the template to learn more about
   the dracut booting process.

.. code-block:: shell

   wwctl node set wwnode1 --tagadd GrubMenuEntry=dracut

.. note::

   The ``GrubMenuEntry`` variable may be set at the node or profile level.

During boot, ``warewulfd`` will detect and dynamically serve an
initramfs from a node's container image in much the same way that it
can serve a kernel from a container image. This image is loaded by
iPXE (or GRUB) which directs dracut to fetch the node's container image
during boot.

The wwinit module provisions to tmpfs. By default, tmpfs is permitted
to use up to 50% of physical memory. This size limit may be adjustd
using the kernel argument `wwinit.tmpfs.size`. (This parameter is
passed to the `size` option during tmpfs mount. See ``tmpfs(5)`` for
more details.)

.. warning::

   Kernel overrides are not currently fully supported during dracut initramfs boot.

