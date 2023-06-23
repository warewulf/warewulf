=====================
Warewulf Installation
=====================

There are multiple methods to install Warewulf. This page describes
the installation process for some of those methods.

Binary RPMs
===========

The Warewulf project builds binary RPMs as part of its CI/CD
process. You can obtain them from the `GitHub releases
<https://github.com/warewulf/warewulf/releases>`_ page.

Rocky Linux 8
-------------

.. code-block:: console

   # dnf install https://github.com/warewulf/warewulf/releases/download/v4.4.0/warewulf-4.4.0-1.git_afcdb21.el8.x86_64.rpm

openSuse Leap
-------------

.. code-block:: console

   # zypper install https://github.com/warewulf/warewulf/releases/download/v4.4.0/warewulf-4.4.0-1.git_afcdb21.suse.lp153.x86_64.rpm

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
binary RPMs. You can obtain them from the `GitHub releases
<https://github.com/hpcng/warewulf/releases>`_ page.

Select the version you wish to install and download the tarball to any
location on the server, then follow these directions making the
appropriate substitutions:

.. code-block:: bash

   # EDIT HERE
   VERSION=4.2.0
   DOWNLOAD=/tmp/warewulf-${4.2.0}.tar.gz

   # COPY/PASTE THIS
   mkdir ~/src
   cd ~/src
   tar xvf ${DOWNLOAD}
   cd warewulf-${VERSION}
   make all && sudo make install

Git
---

Warewulf is developed in "Git", a source code management platform that
allows collaborative development and revision control. From the Git
repository, you can download different versions of the project either
from tags or branches. By default, when you go to the GitHub page, you
will find the default branch entitled ``main``. The ``main`` branch is
where most of the active development occurs, so if you want to obtain
the latest and greatest version of Warewulf, this is where to go. But
be forewarned, using a snapshot from ``main`` is not guaranteed to be
stable or generally supported for production. If you are building for
production, it is best to download a release tarball from the main
site, the GitHub releases page, or from a Git tag.

.. code-block:: bash

   mkdir ~/git
   cd ~/git
   git clone https://github.com/warewulf/warewulf.git
   cd warewulf
   git checkout main # or switch to a tag like '4.2.0'
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
