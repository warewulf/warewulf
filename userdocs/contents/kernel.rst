=================
Kernel Management
=================

Node Kernels
============

Warewulf nodes require a Linux kernel to boot. There are multiple ways
to do this, but the default, and easiest way is to install the kernel
you wish to use for a particular container, into the container.

Warewulf will locate the kernel automatically within the container and
by default use that kernel for any node configured to use that
container image.

You can see what kernel is included in a container by using the
``wwctl container list`` command:

.. code-block:: console

   # wwctl container list
   CONTAINER NAME            NODES  KERNEL VERSION
   alpine                    0
   rocky                     0      4.18.0-348.12.2.el8_5.x86_64
   rocky_updated             1      4.18.0-348.23.1.el8_5.x86_64

Here you will notice the alpine contianer that was imported has no
kernel within it, and each of the rocky containers include a kernel.

This model was introduced in Warewulf 4.3.0. Previously, Warewulf
managed the kernel and the container separately, which made it hard to
build and distribute containers that have custom drivers and/or
configurations included (e.g. OFED, GPUs, etc.).

Kernel Overrides
================

It is still possible to specify a kernel to a container if it doesn't
include it's own kernel, or if you wish to override the default kernel
by using the ``kernel override`` capability.

You can specify this option either within the ``nodes.conf`` directly
or via the command line with the ``--kerneloverride`` option to
``wwctl node set`` or ``wwctl profile set`` commands.

In this case you will also need to import a kernel specifically into
Warewulf for this purpose using the ``wwctl kernel import`` command as
follows:

.. code-block:: console

   # wwctl kernel import $(uname -r)
   4.18.0-305.3.1.el8_4.x86_64: Done

This process will import not only the kernel image itself, but also
all of the kernel modules and firmware associated to this kernel.

Listing All Imported Kernels
----------------------------

Once the kernel has been imported, you can list them all with the
following command:

.. code-block:: console

   # wwctl kernel list
   VNFS NAME                           NODES
   4.18.0-305.3.1.el8_4.x86_64             0

Once a kernel has been imported and showing up in this list you can
configure it to boot compute nodes.
