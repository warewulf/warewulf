=================
Kernel Management
=================

Node Kernels
============

Warewulf nodes require a Linux kernel to boot. There are multiple ways
to do this, but the default and easiest way is to install the kernel
you wish to use for a particular container into the container.

Warewulf will locate the kernel automatically within the container and
by default use that kernel for any node configured to use that
container image.

You can see what kernel is included in a container by using the
``wwctl container list`` command:

.. code-block:: console

   # wwctl container list
     CONTAINER NAME  NODES  KERNEL VERSION                 CREATION TIME        MODIFICATION TIME    SIZE
     alpine          0                                     05 Jun 23 20:02 MDT  05 Jun 23 20:02 MDT  17.9 MiB
     rocky-8         1      4.18.0-372.13.1.el8_6.x86_64   17 Jan 23 23:48 MST  06 Apr 23 09:40 MDT  2.4 GiB

Here you will notice the alpine contianer that was imported has no
kernel within it, and the rocky container includes a kernel.

This model was introduced in Warewulf v4.3. Previously, Warewulf
managed the kernel and the container separately, which made it hard to
build and distribute containers that have custom drivers and/or
configurations included (e.g. OFED, GPUs, etc.).

Kernel Overrides
================

It is still possible to specify a kernel for a container if it doesn't
include it's own kernel, or if you wish to override the default kernel
by using the ``kernel override`` capability.

You can specify this option with the ``--kerneloverride`` option to
``wwctl node set`` or ``wwctl profile set`` commands.

In this case you will also need to import a kernel specifically into
Warewulf for this purpose using the ``wwctl kernel import`` command as
follows:

.. code-block:: console

   # wwctl kernel import $(uname -r)
   4.18.0-305.3.1.el8_4.x86_64: Done

This process will import not only the kernel image itself, but also
all of the kernel modules and firmware associated with this kernel.

Listing All Imported Kernels
----------------------------

Once the kernel has been imported, you can list them all with the
following command:

.. code-block:: console

   # wwctl kernel list
   VNFS NAME                           NODES
   4.18.0-305.3.1.el8_4.x86_64             0

   # wwctl kernel list
   KERNEL NAME                              KERNEL VERSION            NODES
   4.18.0-305.3.1.el8_4.x86_64                                             0

Once a kernel has been imported you can configure it to boot compute
nodes.
