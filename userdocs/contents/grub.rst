===========
Use of Grub
===========

Instead of the iPXE starter a combination of `shim and GRUB <https://www.suse.com/c/uefi-secure-boot-details/>`_
can be used with the advantage that secure boot can be used. That means 
that only the signed kernel of a distribution can be booted. This can
be a huge security benefit for some scenarios.

In order to enable the grub boot method it has to ne enabled in `warewulf.conf`.
Nodes which are not known to warewulf will then booted with the shim/grub from
the host on which warewulf is installed.


Boot process
============

The boot process can be summarized with following diagram

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

If secure boot is enabled at every step a signature is checked and the boot process
will fail if this check fails. Also at moment a Shim only includes the key 
of one Distribution, which means that every Distribution needs a separate
`shim` and `grub` executable and warewulf extracts these binaries from
the containers.

For the case when the node is unknown to warewulf or
can't be identified during the `tFTP`` boot phase, the shim/grub binaries of
the host in which warewulf is running will be used.

PXE/tFTP boot
-------------

The standard network boot process with `grub` and `iPXE` has following steps

.. graphviz::

  digraph G{
      node [shape=box];
      compound=true;
      edge [label2node=true]
      bios [shape=record label="{Bios | boots filename from nextboot per tFTP}"]
      subgraph cluster0 {
       label="iPXE boot"
       iPXE;
       ipxe_cfg [shape=record label="{ipxe.cfg|generated for indivdual node}"];
       iPXE -> ipxe_cfg [label="http"];
      }
      bios->iPXE [lhead=cluster0,label="filename=iPXE.efi"];
      bios->shim [lhead=cluster1,label="filename=shim.efi"];
      subgraph cluster1{
        label="Grub boot"
        shim[shape=record label="{shim.efi|from ww4 host}"];
        grub[shape=record label="{grubx64.efi | name hardcoded in shim.efi|from ww4 host}"]
        shim->grub[label="tFTP"];
        grubcfg[shape=record label="{grub.cfg|static under tFTP root}"];
        grub->grubcfg[label="tFTP"];
      }
      kernel [shape=record label="{kernel|ramdisk (root fs)|wwinit overlay}|extracted from node container"];
      grubcfg->kernel[ltail=cluster1,label="http"];
      ipxe_cfg->kernel[ltail=cluster0,label="http"];
  }

As the tFTP server is independent of warewulf, the `shim` and `grub` EFI binaries
for the tFTP server are copied from the host on which warewulf is running.
This means that for secure boot the distributor e.g. SUSE of the container in
the `default` profile must match the distributor of the container which then
also must be signed by the SUSE key.

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

The main difference is that the initial `shim.efi` and `grub.efi` are delivered by http with warewulf
and are taken directly from the container assigned to the node. This means that secure boot will work 
for containers from different distributors.

Install shim and efi
--------------------

The `shim.efi` and `grub.efi` must be installed via the package manager directly into the container.

Install on SUSE systems
^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: console

  # wwctl container shell leap15.5 
  [leap15.5] Warewulf> zypper install grub2 shim


Install on EL system
^^^^^^^^^^^^^^^^^^^^

.. code-block:: console

  # wwctl container shell rocky9
  [rocky9] Warewulf> dnf install shim-x64.x86_64 grub2-pc.x86_64