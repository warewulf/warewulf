.. _selinux_images:

======================
SELinux-enabled Images
======================

Warewulf supports booting SELinux-enabled images, though nodes using SELinux
must be configured to use tmpfs for their image file system. ("ramfs," often
used by default, does not support extended file attributes, which are required
for SELinux context labeling.)

.. code-block:: bash

   wwctl profile set default --root tmpfs

.. note::

   Versions of Warewulf prior to v4.5.8 also required a kernel argument
   "rootfstype=ramfs" in order for wwinit to copy the node image to tmpfs; but
   this is no longer required.

Once that is done, enable SELinux in ``/etc/sysconfig/selinux`` and install the
appropriate packages in the image. `An example`_ of such an image is available
in the warewulf-node-images repository.

.. _An example: https://github.com/warewulf/warewulf-node-images/tree/main/examples/rockylinux-9-selinux

SELinux requires extended attributes, which aren't supported on a default
``initrootfs``. Nodes using SELinux should specify ``--root=tmpfs``.