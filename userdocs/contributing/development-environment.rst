=======================
Development Environment
=======================

To develop and test the Warewulf server, you need a single system (typically a
virtual machine) to serve as a test server deployment. To actually test
provisioning your development server also needs a dedicated network that it can
run DHCP on. This can typically be provisioned as a virtual network bridge in
virtual machine software.

Options include:

* KVM / Libvirt
* VirtualBox
* VMWare
* UTM

A Warewulf development environment should likely use Rocky Linux 9 or openSUSE
LEAP 15, though there are ongoing development efforts using Debian and Ubuntu as
well.)

Compiling Warewulf for a Development Server
===========================================

.. code-block:: shell

   # Rocky Linux 9
   dnf -y install git epel-release golang {libassuan,gpgme}-devel unzip tftp-server dhcp-server nfs-utils ipxe-bootimgs-{x86,aarch64}

   git clone https://github.com/warewulf/warewulf.git
   cd warewulf
   env \
     PREFIX=/opt/warewulf \
     SYSCONFDIR=/etc \
     IPXESOURCE=/usr/share/ipxe \
     WWPROVISIONDIR=/opt/warewulf/provision \
     WWOVERLAYDIR=/opt/warewulf/overlays \
     WWCHROOTDIR=/opt/warewulf/chroots \
     make all
   make install

These paths balance isolation (e.g., installing binaries in
``/opt/warewulf/bin/``) with integration (e.g., storing configuration in
``/etc/warewulf/`` and using local Dracut and iPXE paths).

After making changes to the source, simply running ``make install`` should be
enough to update installed binaries.

You should likely also disable any local firewall. Otherwise, consult the
general installation guide for configuration details.

.. code-block:: shell

   systemctl disable --now firewalld

Running the Test Suite
======================

Warewulf includes an ever-growing test suite. Alias targets in the ``Makefile``
support running it quickly, easily, and consistently.

.. code-block:: shell

   make test

Additional tests exist as well to perform various checks on the golang source.
These checks are run automatically by GitHub as part of the Warewulf CI process;
but it is a good idea to run them locally before submitting a new PR.

.. code-block:: shell

   make vet
   make staticcheck
   make lint

New code, and code changes, should often be accompanied by updates to the test
suite.

More information:

* `The golang testing package <https://pkg.go.dev/testing>`_
* `Table Driven Tests <https://go.dev/wiki/TableDrivenTests>`_
* `Testift assert <https://pkg.go.dev/github.com/stretchr/testify/assert>`_
* `Warewulf testenv <https://pkg.go.dev/github.com/warewulf/warewulf/internal/pkg/testenv>`_

Using a Dev Container
=====================

Visual Studio Code (VSC) can utilize a Dev Container for a self-contained
environment that has all the necessary tools and dependencies to build and test
Warewulf. The Dev Container is based on the Rocky 9 image and is built using the
`devcontainer.json` file in the `.devcontainer` directory of the Warewulf
repository.  To use this working Docker/Podman and VSC installations are
required.  To use the Dev Container, click the "Open a Remote Window" button on
the bottom left of the editor (`><` icon) and select "Reopen in Container".
This will build the container and open a new VSC window with the container as
the development environment. 

Using Vagrant and Libvirt
=========================

Vagrant can be used to quickly spin up a test/development environment for Warewulf.
A `Vagrantfile` is provided in `vagrant` directory of the Warewulf repository.
See the `README.md <https://github.com/warewulf/warewulf/blob/main/vagrant/README.md>`_
for more details.