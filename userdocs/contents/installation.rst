=====================
Warewulf Installation
=====================

There are multiple methods to install Warewulf. This page describes
the installation process for some of those methods.

Binary RPMs
===========

The Warewulf project builds binary RPMs as part of its CI/CD
process. You can obtain them from the `GitHub releases`_ page.

.. _GitHub releases: https://github.com/warewulf/warewulf/releases

Rocky Linux 8
-------------

.. code-block:: console

   # dnf install https://github.com/warewulf/warewulf/releases/download/v4.5.8/warewulf-4.5.8-1.el8.x86_64.rpm

openSuse Leap
-------------

.. code-block:: console

   # zypper install https://github.com/warewulf/warewulf/releases/download/v4.5.8/warewulf-4.5.8-1.suse.lp155.x86_64.rpm

Container images
================

Warewulf is prepared to be built inside and packaged into a Linux container.
This can be especially useful for testing and development or just replace classic package installation.
It is also possible to only use the container for building and the install it in the host system afterwards.
For that look at the INSTALL, UNINSTALL and PURGE labels inside the `Dockerfile`_

.. _Dockerfile: https://github.com/warewulf/warewulf/blob/main/Dockerfile

Docker
------

.. code-block:: console

   # docker build -t warewulf .
   # docker run -d --replace --name warewulf-test --privileged --net=host -v /:/host -v /etc/warewulf:/etc/warewulf -v /var/lib/warewulf/:/var/lib/warewulf/ -e NAME=warewulf-test -e IMAGE=warewulf warewulf

Systemd-nspawn
--------------

Since Warewulf runs multiple services inside one single container it uses systemd as init system.
Since a full privileged Docker container running a systemd can cause some side effects,
it might be a better option to use `systemd-nspawn`_ in some cases which was explicitly made to run
containers with a full init system.

.. _systemd-nspawn: https://www.freedesktop.org/software/systemd/man/latest/systemd-nspawn.html

.. code-block:: console

   # docker build -t warewulf .
   # mkdir warewulf-nspawn
   # docker export "$(docker create --name warewulf-test warewulf true)" | tar -x -C warewulf-nspawn
   # systemd-nspawn -D warewulf-nspawn/ passwd
   # systemd-nspawn -D warewulf-nspawn/ --boot


Compiled Source code
====================

Before you build the Warewulf source code you will first need to
install the build dependencies:

* ``make``: This should be available via your Linux distribution's
  package manager (e.g. ``dnf install make``)
* ``go``: Golang is also available on most current Linux
  distributions, but if you wish to install the most recent version,
  you can find that here: `https://golang.org/dl/
  <https://golang.org/dl/>`_
* Depending on your Linux Distribution, you may need to install other
  development packages. Typically it is recommended to install the
  entire development group like ``dnf groupinstall "Development
  Tools"``

Once these dependencies are installed, you can obtain and build the
source code as follows:

Release Tarball
---------------

The Warewulf project releases source distributions alongside its
binary RPMs. You can obtain them from the `GitHub releases`_ page.

Select the version you wish to install and download the tarball to any
location on the server, then follow these directions making the
appropriate substitutions:

.. code-block:: bash

   curl -LO https://github.com/warewulf/warewulf/releases/download/v4.5.8/warewulf-4.5.8.tar.gz
   tar -xf warewulf-4.5.8.tar.gz
   cd warewulf-4.5.8
   make all && sudo make install

Git
---

Warewulf is developed in GitHub, a source code management platform
that allows collaborative development and revision control. From the
Git repository, you can download different versions of the project
either from tags or branches. By default, when you go to the GitHub
page, you will find the default branch entitled ``main``. The
``main`` branch is where most of the active development occurs,
so if you want to obtain the latest and greatest version of Warewulf,
this is where to go. But be forewarned, using a snapshot from
``main`` is not guaranteed to be stable or generally supported
for production.

If you are building for production, it is best to download a release
tarball from the main site, the GitHub releases page, or from a Git
tag.

.. code-block:: bash

   git clone https://github.com/warewulf/warewulf.git
   cd warewulf
   git checkout main # or switch to a tag like 'v4.5.8'
   make all && sudo make install

Runtime Dependencies
--------------------

In Warewulf's default configuration, it will require some operating
system provided services. Generally these are provided by your
installation vendor and can be installed over the network.

These are the services you will need to install:

* ``dhcp-server``
* ``tftp-server``
* ``nfs-utils``

If you are using an Enterprise Linux compatible distribution you can
install them with ``yum install dhcp-server tftp-server nfs-utils``.
