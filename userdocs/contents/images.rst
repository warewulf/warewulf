================
Image Management
================

Since the inception of Warewulf over 20 years ago, Warewulf has used
the model of the "Virtual Node File System" (VNFS) as a template image
for the compute nodes. This is similar to a golden master image,
except that the node file system exists within a directory on the
Warewulf control node (e.g. a ``chroot()``).

Meanwhile, the enterprise has created
a lot of tooling and standards around defining, building,
distributing, securing, and managing containers, so Warewulf v4 now
integrates directly within the container ecosystem to facilitate the
process of image management.

If you are not currently leveraging the container ecosystem in any
other way, you can still build your own chroot directories and use
Warewulf as before.

It is important to understand that Warewulf is not running a container
runtime on cluster nodes. While it is absolutely possible to run
containers on cluster nodes, Warewulf is provisioning the
image to the bare metal and booting it. This image will be used as
the base operating system and, by default, it will run entirely in
memory. This means that when you reboot the node, the node retains no
information about Warewulf or how it booted.

Tools
=====

There are different container managment tools available. Docker is
probably the most recognizable one in the enterprise. Podman is
another one that is gaining traction on the RHEL platforms. In HPC,
Apptainer is the most utilized container management tool. You can use
any of these to create and manage the images to be later imported
into Warewulf.

Structure
=========

A Warewulf image is a directory that populates the runtime root file system of a cluster
node. The image source directory must contain a single ``rootfs`` directory which represents the
actual root directory for the image.

.. code-block:: none

  /var/lib/warewulf/chroots/rockylinux-9
  └── rootfs
      ├── afs
      ├── bin -> usr/bin
      ├── boot
      ├── dev
      ├── etc
      ├── home
      ├── lib -> usr/lib
      ├── lib64 -> usr/lib64
      ├── media
      ├── mnt
      ├── opt
      ├── proc
      ├── root
      ├── run
      ├── sbin -> usr/sbin
      ├── srv
      ├── sys
      ├── tmp
      ├── usr
      └── var

Warewulf images are built (e.g., with ``wwctl image build``) into compressed images for
distribution to cluster nodes.

.. code-block:: none

  /var/lib/warewulf/provision/image
  ├── rockylinux-9.img
  └── rockylinux-9.img.gz

Importing Images
================

Warewulf supports importing an image from any OCI compliant
registry. This means you can import from a public registry or from a
private registry.

Here is an example of importing from Docker Hub.

.. code-block:: console

   # wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:8 rocky-8
   Getting image source signatures
   Copying blob d7f16ed6f451 done
   Copying config da2ca70704 done
   Writing manifest to image destination
   Storing signatures
   [LOG]       info unpack layer: sha256:d7f16ed6f45129c7f4adb3773412def4ba2bf9902de42e86e77379a65d90a984
   Updating the image's /etc/resolv.conf
   Building image: rocky-8

.. note::

    Most images in Docker Hub are not "bootable", in that, they
    have a limited version of Systemd to make them lighter weight for
    image purposes. For this reason, don't expect any base Docker
    image (e.g. ``docker://rockylinux`` or ``docker://debian``) to
    boot properly. They will not, as they will get stuck into a single
    user mode. The images in `https://github.com/warewulf/warewulf-node-images
    <https://github.com/warewulf/warewulf-node-images>`_ are not limited and thus
    they boot as you would expect.

Platform
--------

By default,
Warewulf will try to import an image of the same platform
(e.g., amd64, arm64)
as the local system.
To specify the platform to import,
either specify `WAREWULF_OCI_PLATFORM`
or use the argument `--platform` during import.

Private Registry
----------------

It is possible to use a private registry that is password protected or
does not have the requirement for TLS. In order to do so, you have two
choices for handling the credentials.

* Set environmental variables
* Use ``docker login`` or ``podman login`` which will store the
  credentials locally

Please note, there is no requirement to install and use docker or
podman on your control node just for importing images into Warewulf.

Here are the environmental variables that can be used.

.. code-block:: console

   WAREWULF_OCI_USERNAME
   WAREWULF_OCI_PASSWORD
   WAREWULF_OCI_NOHTTPS

They can be overwritten with ``--nohttps``, ``--username`` and ``--password``.
.. code-block:: console

   # wwctl import --username tux --password supersecret docker://ghcr.io/privatereg/rocky:8

The above is just an example. Consideration should be done before
doing it this way if you are in a security sensitive environment or
shared environments as this command line wil show up in the process 
table.

Local Files
-----------

It is also possible to import an image from a local file or
directory. For example, Podman can save a `.tar` archive of an OCI
image. This archive can be directly imported into Warewulf, no
registry required.

.. code-block:: console

   # podman save alpine:latest >alpine.tar
   # wwctl image import alpine.tar alpine

Chroot directories and Apptainer sandbox images can also be imported
directly.

.. code-block:: console

   $ apptainer build --sandbox ./rockylinux-8/ docker://ghcr.io/warewulf/warewulf-rockylinux:8
   $ sudo wwctl image import ./rockylinux-8/ rockylinux-8

HTTP proxies
------------

You can set ``HTTP_PROXY``, ``HTTPS_PROXY``, and ``NO_PROXY`` (or their
lower-case versions) to use a proxy during ``wwctl image import``.

.. code-block:: shell

   export HTTPS_PROXY=squid.localdomain
   wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:8

See ProxyFromEnvironment_ For more information.

.. _ProxyFromEnvironment: https://pkg.go.dev/net/http#ProxyFromEnvironment

.. note::

   OCI and ORAS registries typically use HTTPS, so you probably need to set
   ``HTTPS_PROXY`` or ``https_proxy`` rather than the ``HTTP`` variants.

Syncuser
========

At import time Warewulf checks if the names of the users on the host
match the users and UIDs/GIDs in the imported image. If there is
mismatch, the import command will print out a warning.  By setting the
``--syncuser`` flag you advise Warewulf to try to syncronize the users
from the host to the image, which means that ``/etc/passwd`` and
``/etc/group`` of the imported image are updated and all the files
belonging to these UIDs and GIDs will also be updated.

A check if the users of the host and image matches can be
triggered with the ``syncuser`` command.

.. code-block:: console

   # wwctl image syncuser image-name

With the ``--write`` flag it will update the image to match the
user database of the host as described above.

.. code-block:: console

   wwctl image syncuser --write image-name

Listing All Imported Images
===========================

Once the image has been imported, you can list them all with the
following command:

.. code-block:: console

   # wwctl image list
   IMAGE NAME
   ----------
   rocky-8

Once an image has been imported and showing up in this list you can
configure it to boot compute nodes.

Making Changes To Images
========================

You can run commands inside of any of the images and make changes to
them as follows:

.. code-block:: console

   # wwctl image exec rocky-8 /bin/sh
   [rocky-8] Warewulf> cat /etc/rocky-release
   Rocky Linux release 8.4 (Green Obsidian)
   [rocky-8] Warewulf> exit
   Rebuilding image...
   [INFO]     Skipping (image is current)

You can also ``--bind`` directories from your host into the image
when using the exec command. This works as follows:

.. code-block:: console

   # wwctl image shell --bind /tmp:/mnt rocky-8
   [rocky-8] Warewulf>

.. note::

   As with any mount command, both the source and the target must
   exist. This is why the example uses the ``/mnt/`` directory
   location, as it is almost always present and empty in every Linux
   distribution (as prescribed by the LSB file hierarchy standard).

Files which should always be present in an image like ``resolv.conf``
can be specified in ``warewulf.conf``:

.. code-block:: yaml

   image mounts:
   - source: /etc/resolv.conf
     dest: /etc/resolv.conf
     readonly: true

.. note::

   Instead of ``readonly: true`` you can set ``copy: true``. This causes the
   source file to be copied to the image and removed if it was not
   modified. This can be useful for files used for registrations.

When the command completes, if anything within the image changed,
the image will be rebuilt into a bootable static object
automatically. (To skip the automatic image rebuild, specify ``--build=false``.)

If the files ``/etc/passwd`` or ``/etc/group`` were updated, there
will be an additional check to confirm if the users are in sync as
described in `Syncuser`_ section.

Excluding Files from an Image
-----------------------------

Warewulf can exclude files from an image source to prevent them
from being delivered to the compute node. This is typically used to
reduce the size of the image when some files are unnecessary.

Patterns for excluded files are read from the file
``/etc/warewulf/excludes`` in the image itself. For example,
the default Rocky Linux images exclude these paths:

.. code-block::

   /boot/
   /usr/share/GeoIP

``/etc/warewulf/excludes`` supports the patterns implemented by
`filepath.Match <https://pkg.go.dev/path/filepath#Match>`_.

Preparing an image for build
----------------------------

Warewulf executes the script ``/etc/warewulf/image_exit.sh`` after
a ``wwctl image shell`` or ``wwctl image exec`` and prior to
(re)building the final node image for delivery. This is typically used
to remove cache or log files that may have been generated by the
executed command or interactive session.

For example, the default Rocky Linux images runs ``dnf clean all`` to
remove any package repository caches that may have been generated.

Creating Images From Scratch
============================

It is absolutely possible to create an `OCI base image`_ from scratch, but it is
particularly easy to do with Apptainer.

.. _OCI base image: https://docs.docker.com/build/building/base-images/

Consider the following file called `warewulf-rockylinux-9.def`:

.. code-block:: singularity

   Bootstrap: yum
   MirrorURL: https://download.rockylinux.org/pub/rocky/9/BaseOS/x86_64/os/
   Include: dnf

   %post
   dnf -y install --allowerasing \
     NetworkManager \
     basesystem \
     bash \
     curl-minimal \
     kernel \
     nfs-utils \
     openssh-server \
     systemd

   dnf -y remove \
     glibc-gconv-extra
   rm -rf /boot/* /run/*
   dnf clean all

Warewulf cannot directly import a container image from an Apptainer SIF yet, so
an Apptainer image must be built as a *sandbox*.

.. code-block:: console

   # apptainer build --sandbox warewulf-rockylinux-9 warewulf-rockylinux-9.def
   [...]
   INFO:    Creating sandbox directory...
   INFO:    Build complete: warewulf-rockylinux-9

Once a sandbox container image has been built, it can be imported into Warewulf.

.. code-block:: console

   # wwctl container import ./warewulf-rockylinux-9 rockylinux-9

.. note::

   Although warewulf does not currently support importing a SIF directly, a SIF can be converted to
   a sandbox with Apptainer and then imported into Warewulf.
    
   .. code-block:: console

      # apptainer build --sandbox my-sandbox my-image.sif
      # wwctl container import ./my-sandbox my-image

Duplicating an image
====================

It is possible to duplicate an installed image by using:

.. code-block:: console

  # wwctl image copy IMAGE_NAME DUPLICATED_IMAGE_NAME

This kind of duplication can be useful if you are looking for canary tests.

.. note::

   If an image source includes persistent sockets, these sockets may cause the copy operation to fail.

   .. code-block:: console

      Copying sources...
      ERROR  : could not duplicate image: lchown /var/lib/warewulf/chroots/rocky-8/rootfs/run/user/0/gnupg/d.kg8ijih5tq41ixoeag4p1qup/S.gpg-agent: no such file or directory

   To resolve this, remove the sockets from the image source.

   .. code-block:: bash

      find $(wwctl image show rocky-8) -type s -delete

Multi-arch image management
===========================

It is possible to build, edit, and provision images of different
architectures (i.e. aarch64) from an x86_64 host by using QEMU. Simply 
run the appropriate command below based on your image management tools.

.. code-block:: console

   # sudo docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
   # sudo podman run --rm --privileged multiarch/qemu-user-static --reset -p yes
   # sudo singularity run docker://multiarch/qemu-user-static --reset -p yes

Then, ``wwctl image exec`` will work regardless of the architecture of the image.
For more information about QEMU, see their `GitHub <https://github.com/multiarch/qemu-user-static>`_

To use wwclient on a booted image using a different architecture,
wwclient must be compiled for the specific architecture. This requires GOLang build
tools 1.21 or newer. Below is an example for building wwclient for arm64:

.. code-block:: console

   # git clone https://github.com/warewulf/warewulf
   # cd warewulf
   # GOARCH=arm64 PREFIX=/ make wwclient
   # mkdir -p /var/lib/warewulf/overlays/wwclient_arm64/rootfs/warewulf
   # cp wwclient /var/lib/warewulf/overlays/wwclient_arm64/rootfs/warewulf

Then, apply the new "wwclient_arm64" system overlay to your arm64 node/profile

Read-only images
================

An image may be marked "read-only" by creating a ``readonly`` file in its
source directory, typically next to ``rootfs``.

.. note::

   Read-only images are a preview feature primarily meant to enable future
   support for image subscriptions and updates.
