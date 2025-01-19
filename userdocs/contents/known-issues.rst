============
Known issues
============

SELinux and IPMI write not working when using two-stage boot
============================================================

The dracut implementation of two-stage boot in versions of Warewulf prior to v4.6.0 bypasses the
``wwinit`` process by default, invoking the image's init system directly. While cluster nodes will
often still boot mostly successfully this way, features implemented by wwinit will not complete. In
particular, SELinux relabeling and IPMI write are not executed.

To ensure that dracut runs the full ``wwinit`` process, pass ``init=/init`` or
``init=/warewulf/wwinit`` on the kernel command line.

.. code-block:: bash

   # wwctl profile set default --kernelargs="init=/init"

Images are read-only
====================

Warewulf v4.5 uses the permissions on an image's ``rootfs/`` to determine a "read-only" state of
the image: if the root directory of the image is ``u-w``, it will be mounted read-only
during ``wwctl image <exec|shell``, preventing interactive changes to the image.

In the past, the root directory was ``u+w``, but Enterprise Linux 9.5 (including Red Hat, Rocky, _et
al._) includes an update to the ``filesystem`` package that marks the root directory ``u-w``. This
causes Warewulf images to be "read only" by default.

To mark a Warewulf image as writeable, use `chmod u+w`.

.. code-block:: bash

   # chmod u+w $(wwctl image show rockylinux-9.5)

This behavior is changed in v4.6 to use an explicit ``readonly`` file stored outside of ``rootfs/``.

Image sockets cause build failures
==================================

If an image source directory includes persistent sockets, these sockets may cause the import operation to fail.

.. code-block:: console

   Copying sources...
   ERROR  : could not import image: lchown ./rockylinux-8/run/user/0/gnupg/d.kg8ijih5tq41ixoeag4p1qup/S.gpg-agent: no such file or directory

To resolve this, remove the sockets from the source directory.

.. code-block:: bash

   find ./rockylinux-8/ -type s -delete

This issue was fixed in an upstream library and `should be resolved in Warewulf
v4.6.0. <https://github.com/warewulf/warewulf/issues/892>`_
