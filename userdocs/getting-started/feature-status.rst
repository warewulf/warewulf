==============
Feature Status
==============

The following table summarizes the maturity of Warewulf's top-level features.
Features are classified as follows:

* **Stable** -- Production-ready; the interface is well-tested and unlikely to
  change in incompatible ways.
* **Preview** -- Functional, but may have rough edges, limited testing, or an
  interface that is still evolving.
* **Incomplete** -- Work-in-progress; key functionality or the user-facing
  interface is not yet complete.

.. list-table::
   :widths: 30 15 55
   :header-rows: 1

   * - Feature
     - Status
     - Notes
   * - iPXE
     - Stable
     - Default network bootloader used for PXE booting cluster nodes.
   * - GRUB
     - Preview
     - Alternative bootloader; useful for UEFI and Secure Boot scenarios.
   * - Single-stage boot
     - Stable
     - Default provisioning mode; the OS image is used directly as the root
       filesystem.
   * - Two-stage boot
     - Preview
     - Uses an intermediate initrd to initialize hardware before loading the
       final OS image. Requires Dracut-based images.
   * - Nodes
     - Stable
     - Core node management (add, remove, configure, list).
   * - Profiles
     - Stable
     - Abstract node profiles for sharing configuration across groups of nodes.
   * - Overlays
     - Stable
     - Template-based file provisioning applied to OS images at boot and at
       runtime.
   * - OS images
     - Stable
     - Container-based OS image import, management, and provisioning.
   * - Read-only images
     - Preview
     - Marking an image read-only to enable future support for image
       subscriptions and updates.
   * - Kernels
     - Stable
     - Kernel extraction from OS images and per-node kernel management.
   * - Disk provisioning
     - Preview
     - Partitioning and formatting disks on cluster nodes during provisioning.
   * - Provision-to-disk
     - Preview
     - Provisioning an OS image to disk so that the node can subsequently boot
       without Warewulf.
   * - Resources
     - Incomplete
     - Generic node resources; the user-facing interface is not yet complete.
   * - Secure Boot
     - Preview
     - UEFI Secure Boot support via signed bootloaders.
   * - dnsmasq
     - Preview
     - Using dnsmasq as an alternative to ISC dhcpd and TFTP for DHCP and
       network boot services.
