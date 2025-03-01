====================
Cluster Provisioning
====================

Clusters have many scalability factors to consider. Often overlooked among them
is "administrative scaling"-- the systems administration overhead of a person or
team maintaining a large number of systems. While homogeneous configurations do
improve administrative scaling, each installed server is still subject to
version and configuration drift, eventually becoming a point of discrete
administration and debugging. The larger the cluster, the harder this problem is
to solve.

This is the problem that Warewulf was created solve.

Provisioning Overview
=====================

Provisioning is the process of preparing a system for use, typically by
providing and configuring an operating system. There are many ways to accomplish
this, from copying hard drives, to scripted installs, to automated installs.
Each has its place, and there are many tools available to facilitate each
method.

Before dedicated cluster provisioning systems, administrators would visit each
cluster node and install it from scratch, with an ISO, CD, or USB flash drive.
This is obviously not scalable. Because the nodes in a cluster environment are
typically identical, it is much more efficient to group sets of nodes together
to be provisioned in bulk.

Why Stateless Provisioning
==========================

Warewulf further improves on the automated provisioning process by skipping the
installation completely; it boots directly into the runtime operating system
without ever doing an installation.

Stateless provisioning means you never have to install another compute node.
Think of it like booting a LiveOS or LiveISO on nodes over the network. This
means that no node requires discrete administration, but rather the entire
cluster is administrated as a single unit. There is no version drift, because it
is not possible for nodes to fall out of sync. Every reboot makes it exactly the
same as its neighbors.

Cluster Node Requirements
=========================

The only requirement to provision a node with Warewulf is that the node is set
to PXE boot. You may need to change the boot order if there is a local disk
present and bootable. This is a configuration change you will have to make in
the BIOS of the cluster node.

This configuration is different for each vendor platform. For more information,
consult your system documentation or contact your hardware vendor support.

.. note::

   Hardware vendors are often able to preconfigure your cluster nodes with
   values of your choosing. Ask them to provide a text file that includes all of
   the network interface MAC addresses of the clusters nodes in the order they
   are racked--this simplifies the process of adding nodes to Warewulf.

The Provisioning Process
========================

When a cluster node boots from Warewulf, the following process occurs:

#. The system firmware (either BIOS or UEFI) initializes hardware, including
   local network interfaces.
#. The system uses an in-firmware PXE client to obtain a BOOTP/DHCP address from
   the network.
#. The DHCP server (hosted either on the Warewulf server or externally) responds
   with an address suitable for provisioning, along with a "next-server" option
   directing the cluster node to download (via TFTP) and execute a bootloader
   (either iPXE or GRUB) with a Warewulf-provided configuration.
#. The bootloader configuration directs the cluster node to download and
   bootstrap the configured kernel, image, and overlays from the Warewulf (HTTP)
   server.

   * In a single-stage provisioning configuration, the desired image and
     overlays are combined and provisioned immediately by the bootloader as the
     kernel's initial root file system. This is straightfoward, but does not
     work in all environments: some systems have memory layouts that are not
     handled properly by either iPXE or GRUB for sufficiently large image sizes,
     leading to strange, unpredictable results.
   * In a two-stage provisioning configuration, a small initial root fs (created
     by dracut) is provisioned first, and this image uses the provisioned Linux
     kernel to retrieve and deploy the full image and overlays. Perhaps
     counter-intuitively, the two-stage provisioning process is often quicker
     than the single-stage process, because the Linux environment is more I/O
     efficient than the bootloader itself.

#. Optionally included in a configured overlay, ``wwclient`` is left resident on
   the cluster node and periodically refreshes configured runtime overlays.
