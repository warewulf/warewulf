.. _images:

===========
Node Images
===========

Warewulf node images are a "Virtual Node File System" (VNFS) that serves as a
base image for cluster nodes. This is similar to a "golden master" image, except
that the image source exists mutably within a directory on the Warewulf control
node (e.g. a ``chroot()``).

Warewulf node images have several similarities to Linux containers; so Warewulf
v4 integrates directly within the container ecosystem to facilitate the process
of image creation and image management: images can be built, for example, with
Docker, Podman, or Apptainer, and imported directly from OCI registries or local
container image archives. But you can also still build your own chroot
directories manually.

Structure
=========

A Warewulf image is a directory that populates the base runtime root file system
of a cluster node. The image source directory must contain a single ``rootfs``
directory which represents the actual root directory for the image.

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

Importing Images
================

Before any cluster nodes can be provisioned, you must import an image. Images
may be imported from an OCI registry, a local OCI archive, or a local directory
or Apptainer sandbox.

OCI Registry
------------

You can import node images from an OCI registry, public or private.

.. code-block:: console

   # wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:8 rockylinux-8
   Getting image source signatures
   Copying blob d7f16ed6f451 done
   Copying config da2ca70704 done
   Writing manifest to image destination
   Storing signatures
   [LOG]       info unpack layer: sha256:d7f16ed6f45129c7f4adb3773412def4ba2bf9902de42e86e77379a65d90a984
   Updating the image's /etc/resolv.conf
   Building image: rockylinux-8

.. note::

    Most images in Docker Hub are not "bootable": they typically do not include
    a kernel, and likely don't include any init system. For this reason, don't
    expect a base image from DockerHub (e.g. ``docker://rockylinux`` or
    ``docker://debian``) to boot properly with Warewulf.
    
    The Warewulf project maintains a set of `example node images
    <https://github.com/warewulf/warewulf-node-images>`_ that are configured to
    boot when used with Warewulf. These images can be imported directly into
    Warewulf or used as base images for local custom image.

A few environmental variables can be used to control communication with the OCI
registry:

.. code-block:: console

   WAREWULF_OCI_USERNAME
   WAREWULF_OCI_PASSWORD
   WAREWULF_OCI_NOHTTPS

They can be overwritten with ``--nohttps``, ``--username`` and ``--password``.

.. code-block:: console

   # wwctl import --username tux --password supersecret docker://ghcr.io/privatereg/rocky:8

You can also set ``HTTP_PROXY``, ``HTTPS_PROXY``, and ``NO_PROXY`` (or their
lower-case versions) to use a proxy during ``wwctl image import``.

.. code-block:: shell

   export HTTPS_PROXY=squid.localdomain
   wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:8

See ProxyFromEnvironment_ for more information.

.. _ProxyFromEnvironment: https://pkg.go.dev/net/http#ProxyFromEnvironment

.. note::

   OCI and ORAS registries typically use HTTPS, so you probably need to set
   ``HTTPS_PROXY`` or ``https_proxy`` rather than the ``HTTP`` variants.

The above is just an example. Consideration should be done before doing it this
way if you are in a security sensitive environment or shared environments as
this command line wil show up in the process table.

Local OCI Archive
-----------------

It is also possible to import an image from a local OCI archive. For example,
Podman can save a ``.tar`` archive of an OCI image.

.. code-block:: shell

   podman save ghcr.io/warewulf/warewulf-rockylinux:8 >rockylinux-8.tar
   wwctl image import rockylinux-8.tar rockylinux-8

Local Directories and Apptainer Sandboxes
-----------------------------------------

Chroot directories and Apptainer sandbox images can also be imported directly.

.. code-block:: shell

   apptainer build --sandbox ./rockylinux-8/ docker://ghcr.io/warewulf/warewulf-rockylinux:8
   wwctl image import ./rockylinux-8/ rockylinux-8

Listing Imported Images
=======================

Once the image has been imported, you can list them all with ``wwctl image
list``.

.. code-block:: console

   # wwctl image list
   IMAGE NAME
   ----------
   rockylinux-8

Additional detail is available using ``wwctl image list --long``, among others.
(See ``--help`` for more options.)

.. code-block:: console

   # wwctl image list --long
   IMAGE NAME    NODES  KERNEL VERSION      CREATION TIME        MODIFICATION TIME    SIZE
   ----------    -----  --------------      -------------        -----------------    ----
   rockylinux-8  0      4.18.0-553.30.1     11 Feb 25 13:57 MST  11 Feb 25 13:57 MST  1.4 GiB

Modifying Images Interactively 
==============================

An image that has been imported into Warewulf remains mutable, and can be
modified on the Warewulf server. For example, you can "shell" into the image and
make changes interactively.

.. code-block:: console

   # wwctl image shell rockylinux-8
   [warewulf:rockylinux-8] /# dnf -y install apptainer
   [...]

   Installed:
     apptainer-1.3.6-1.el8.aarch64
     fakeroot-1.33-1.el8.aarch64
     fakeroot-libs-1.33-1.el8.aarch64
     fuse3-libs-3.3.0-19.el8.aarch64
     lzo-2.08-14.el8.aarch64
     squashfs-tools-4.3-21.el8.aarch64

   Complete!

Binding Files and Directories
-----------------------------

You can ``--bind`` directories from the Warewulf server into the image when
using the exec command. This is particularly useful for installing locally-built
packages.

.. code-block:: shell

   # wwctl image shell --bind /var/lib/mock/rocky+epel-9-$(arch)/result:/mnt
   [warewulf:rockylinux-8] /# dnf -y install /mnt/warewulf-dracut-*.noarch.rpm

.. note::

   As with any mount command, both the source and the target must exist. This is
   why the example uses the ``/mnt/`` directory location, as it is almost always
   present and empty in every Linux distribution (as prescribed by the LSB file
   hierarchy standard).

Files may also be automatically bound into the image during ``wwctl image
shell`` by configuring ``warewulf.conf:image mounts``.

.. code-block:: yaml

   image mounts:
   - source: /etc/resolv.conf
     dest: /etc/resolv.conf
     readonly: true

.. note::

   Instead of ``readonly: true`` you can set ``copy: true``. This causes the
   source file to be copied to the image and removed if it was not modified.
   This can be useful for files used for registrations.

When the command completes, if anything within the image changed, the image will
be rebuilt into a bootable static object automatically. (To skip the automatic
image rebuild, specify ``--build=false``.)

If the files ``/etc/passwd`` or ``/etc/group`` were updated, there will be an
additional check to confirm if the users are in sync as described in the
:ref:`Syncuser` section.

Specifying a Prompt
-------------------

Warewulf sets a custom prompt during a ``wwctl image shell`` session. This
prompt may be customized using the ``WW_PS1`` variable, which is used to
construct the final ``PS1`` variable for the shell.

.. code-block:: console

   # export WW_PS1="\u@\h:\w\$ "
   # wwctl image shell rockylinux-8
   [warewulf:rockylinux-8] root@rocky:/$

Shell History
-------------

By default, Warewulf image shell sessions don't retain history; but you can
specify a history file by specifying ``WW_HISTFILE``. Note that this file is
stored within the image; you may want to :ref:`exclude` it when the image is
built.

Running Specific Commands
-------------------------

A single command can also be executed in an image, as an alternative to an
interactive shell.

.. code-block:: shell

   wwctl image exec rockylinux-8 -- /usr/bin/dnf -y install apptainer

Building Images
===============

Warewulf images must be built (e.g., with ``wwctl image build``) into compressed
images for distribution to cluster nodes during provisioning.

.. code-block:: console

   # wwctl image build rockylinux-9
   Building image: rockylinux-9
   Created image for Image rockylinux-9: /var/lib/warewulf/provision/images/rockylinux-9.img
   Compressed image for Image rockylinux-9: /var/lib/warewulf/provision/images/rockylinux-9.img.gz

.. _exclude:

Excluding Files
---------------

Warewulf can exclude files from an image to prevent them from being delivered to
the compute node. This is typically used to reduce the size of the image when
some files are unnecessary.

Patterns for excluded files are read from the file ``/etc/warewulf/excludes`` in
the image itself. For example, the default Rocky Linux images exclude these
paths:

.. code-block::

   /boot/
   /usr/share/GeoIP

``/etc/warewulf/excludes`` supports the patterns implemented by `filepath.Match
<https://pkg.go.dev/path/filepath#Match>`_.

Exit Script
-----------

Warewulf executes the script ``/etc/warewulf/image_exit.sh`` in the image after
a ``wwctl image shell`` or ``wwctl image exec`` and prior to (re)building the
final node image for delivery. This is typically used to remove cache or log
files that may have been generated by the executed command or interactive
session.

For example, the default Rocky Linux images runs ``dnf clean all`` to remove any
package repository caches that may have been generated.

Defining New Images
===================

It is absolutely possible to import a base image into Warewulf and make all
changes interactively with ``wwctl image shell``; but it is often better to
define new images with a container image definition file. This can be done using
the OCI and Singularity (Apptainer) ecoystems.

Podman
------

An OCI Containerfile can build from an existing container image to add local
customizations.

.. code-block::

   FROM ghcr.io/warewulf/warewulf-rockylinux:9

   RUN dnf -y install epel-release \
       && dnf -y install apptainer

.. code-block:: console

   # podman build . --file Containerfile --tag custom-image
   [...]
   Successfully tagged localhost/custom-image:latest

   # wwctl image import $(podman image mount localhost/custom-image) custom-image
   # podman image unmount localhost/custom-image

Apptainer
---------

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

   Although warewulf does not currently support importing a SIF directly, a SIF
   can be converted to a sandbox with Apptainer and then imported into Warewulf.
    
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

   If an image source includes persistent sockets, these sockets may cause the
   copy operation to fail.

   .. code-block:: console

      Copying sources...
      ERROR  : could not duplicate image: lchown /var/lib/warewulf/chroots/rocky-8/rootfs/run/user/0/gnupg/d.kg8ijih5tq41ixoeag4p1qup/S.gpg-agent: no such file or directory

   To resolve this, remove the sockets from the image source.

   .. code-block:: bash

      find $(wwctl image show rocky-8) -type s -delete

Image Architecture
==================

By default, Warewulf will try to import an image of the same platform (e.g.,
amd64, arm64) as the local system. To specify the platform to import, either
specify `WAREWULF_OCI_PLATFORM` or use the argument `--platform` during import.

It is possible to build, edit, and provision images of different architectures
(i.e. aarch64) from an x86_64 host by using QEMU. Simply run the appropriate
command below based on your image management tools.

.. code-block:: console

   # docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
   # podman run --rm --privileged multiarch/qemu-user-static --reset -p yes
   # apptainer run docker://multiarch/qemu-user-static --reset -p yes

Then, ``wwctl image exec`` will work regardless of the architecture of the
image. For more information about QEMU, see their `GitHub
<https://github.com/multiarch/qemu-user-static>`_

.. note::

   When provisioning cluster nodes with a different architecture than the
   Warewulf server, also use the matching architecture-specific :ref:`wwclient`
   overlay: e.g., wwclient.x86_64 or wwclient.aarch64.

Read-only images
================

An image may be marked "read-only" by creating a ``readonly`` file in its source
directory, typically next to ``rootfs``.

.. note::

   Read-only images are a preview feature primarily meant to enable future
   support for image subscriptions and updates.
