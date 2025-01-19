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

Image Size Considerations
=========================

Node images can grow quickly as packages and other files are added
to them. Even these larger images are often not an issue in modern
environments; but some architectural limits exist that can impede the
use of images larger than a few gigabytes. Workarounds exist for these
issues in most circumstances:

* Warewulf's :ref:`two-stage boot support <booting with dracut>` effectively
  eliminates this problem by handling the bulk of the image management within
  Linux. This feature is currently in preview, and is subject to change; but it
  is likely to become the default boot method in a future release.

* Systems booting in legacy / BIOS mode, being a 32-bit environment,
  cannot boot an image that requires more than 4GB to decompress. This
  means that the compressed image and the decompressed image together
  must be < 4GB. This is typically reported by the system as "No space
  left on device (https://ipxe.org/34182006)."

  The best work-around for this limitation is to switch to UEFI. UEFI
  is 64-bit and should support booting significantly larger images,
  though sometimes system-specific implementation details have led to
  artificial limitations on image size.

* The Linux kernel itself can only decompress an image up to 4GB due
  to the use of 32-bit integers in critical sections of the kernel
  initrd decompression code.

  The best work-around for this limitation is to use an iPXE with
  support for `imgextract <https://ipxe.org/cmd/imgextract>`_. This
  allows iPXE to decompress the image rather than the kernel.

* Some BIOS / firmware retain a "memory hole" feature for legacy
  devices, e.g., reserving a 1MB block of memory at the 15MB-16MB
  address range. this feature can interfere with booting stateless
  node images.

  If you are still getting "Not enough memory" or "No space left on
  device" errors, try disabling any "memory hole" features or updating
  your system BIOS or firmware.
