=================
Kernel Management
=================

Warewulf nodes require a Linux kernel to boot. As of Warewulf v4.6, the kernel
you wish to use must be present in the relevant image. Warewulf locates and
provisions the kernel automatically for any node configured to use that
image.

You can see what kernels are available in imported images by using the
``wwctl image kernels`` command:

.. code-block:: console

   # wwctl image kernels
   Image                Kernel                                              Version          Default  Nodes
   -----                ------                                              -------          -------  -----
   newroot-test         /boot/vmlinuz-5.14.0-427.37.1.el9_4.aarch64         5.14.0-427.37.1  true     0
   newroot-test         /lib/modules/5.14.0-427.37.1.el9_4.aarch64/vmlinuz  5.14.0-427.37.1  false    0
   rocky-8              /boot/vmlinuz-4.18.0-372.13.1.el8_6.x86_64          4.18.0-372.13.1  true     2
   rocky-8              /lib/modules/4.18.0-372.13.1.el8_6.x86_64/vmlinuz   4.18.0-372.13.1  false    0
   rocky-9.3            /lib/modules/5.14.0-362.13.1.el9_3.aarch64/vmlinuz  5.14.0-362.13.1  true     0
   rockylinux-9-custom  /lib/modules/5.14.0-427.40.1.el9_4.aarch64/vmlinuz  5.14.0-427.40.1  true     0

Kernel Version
==============

If an image includes multiple kernels, the desired kernel may be selected by
specifying the desired version or an explicit path.

.. code-block:: console

   # wwctl node set n1 --kernelversion=4.18.0-372.13.1
   # wwctl node set n1 --kernelversion=/boot/vmlinuz-4.18.0-372.13.1.el8_6.x86_64
