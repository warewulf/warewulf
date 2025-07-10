===================
Server Installation
===================

There are multiple methods to install a Warewulf server. This page describes
some of those methods.

Binary RPMs
===========

The Warewulf project builds binary RPMs as part of its CI/CD process. You can
obtain them from the `GitHub releases`_ page.

.. _GitHub releases: https://github.com/warewulf/warewulf/releases

Rocky Linux 9
-------------

.. code-block:: console

   # dnf install https://github.com/warewulf/warewulf/releases/download/v4.6.2/warewulf-4.6.2-1.el9.x86_64.rpm

openSuse Leap
-------------

.. code-block:: console

   # zypper install https://github.com/warewulf/warewulf/releases/download/v4.6.2/warewulf-4.6.2-1.suse.lp155.x86_64.rpm

Container images
================

Warewulf can be built in a Linux container. This can be especially useful for
testing and development, or to replace traditional package installation. It is
also possible to only use the container for building and the install it in the
host system afterwards. For that look at the INSTALL, UNINSTALL and PURGE labels
inside the `Dockerfile`_

.. _Dockerfile: https://github.com/warewulf/warewulf/blob/main/Dockerfile

Docker
------

.. code-block:: console

   # docker build -t warewulf .
   # docker run -d --replace --name warewulf-test --privileged --net=host -v /:/host -v /etc/warewulf:/etc/warewulf -v /var/lib/warewulf/:/var/lib/warewulf/ -e NAME=warewulf-test -e IMAGE=warewulf warewulf

Systemd-nspawn
--------------

Warewulf runs multiple services inside one single container and uses systemd as
init system. As such, it might be better to use `systemd-nspawn`_, which was
explicitly made to run containers with a full init system.

.. _systemd-nspawn: https://www.freedesktop.org/software/systemd/man/latest/systemd-nspawn.html

.. code-block:: console

   # docker build -t warewulf .
   # mkdir warewulf-nspawn
   # docker export "$(docker create --name warewulf-test warewulf true)" | tar -x -C warewulf-nspawn
   # systemd-nspawn -D warewulf-nspawn/ passwd
   # systemd-nspawn -D warewulf-nspawn/ --boot

Compiled from Source
====================

Before you build the Warewulf source code you will first need to install the
build dependencies:

* ``make``: This should be available via your Linux distribution's package
  manager (e.g. ``dnf install make``)

* ``go``: Golang is also available on most current Linux distributions, but you
  can also install `the most recent version. <https://golang.org/dl/>`_

* Depending on your Linux Distribution, you may need to install other
  development packages. Typically it is recommended to install the entire
  development group.
  
  .. code-block::
   
     dnf groupinstall "Development Tools"

Once these dependencies are installed, you can obtain and build the source code.

Release Tarball
---------------

The Warewulf project releases source distributions alongside its binary RPMs.
You can obtain them from the `GitHub releases`_ page.

Select the version you wish to install and download the tarball to any
location on the server, then follow these directions making the
appropriate substitutions:

.. code-block:: bash

   curl -LO https://github.com/warewulf/warewulf/releases/download/v4.6.2/warewulf-4.6.2.tar.gz
   tar -xf warewulf-4.6.2.tar.gz
   cd warewulf-4.6.2
   make all && sudo make install

Git
---

You can install different versions of Warewulf from its Git tags or branches.
The ``main`` branch is where most active development occurs, so if you want to
obtain the latest and greatest version of Warewulf, this is where to go. But be
forewarned, using a snapshot from ``main`` is not guaranteed to be stable or
generally supported for production.

If you are building for production, it is best to download a release tarball
from the main site, the GitHub releases page, or from a Git tag.

.. code-block:: bash

   git clone https://github.com/warewulf/warewulf.git
   cd warewulf
   git checkout main # or switch to a tag like 'v4.6.2'
   make all && sudo make install

Runtime Dependencies
--------------------

In its default configuration, Warewulf requires some operating system provided
services. Generally these are provided by your distribution.

* ``dhcp-server``
* ``tftp-server``
* ``nfs-utils``

If you are using an Enterprise Linux compatible distribution you can install
them with ``dnf install dhcp-server tftp-server nfs-utils``.

Building RPM packages from source
=================================

You can also build RPM packages from source.

.. code-block:: bash

   dnf -y install epel-release
   dnf -y install make mock
   git clone git@github.com:warewulf/warewulf.git
   (
      cd warewulf
      make clean && make dist warewulf.spec && mock -r rocky+epel-9-$(arch) --rebuild --spec=warewulf.spec --sources=.
   )
   dnf -y install /var/lib/mock/rocky+epel-9-$(arch)/result/warewulf-*.$(arch).rpm

Starting warewulfd
==================

The Warewulf installation registers the Warewulf service with systemd, so it
should be as easy to start/stop/check as any other systemd service:

.. code-block:: console

   # systemctl enable --now warewulfd
