============================
Building RPMs from Source
============================

Install Dependencies
================

.. code-block:: bash

   dnf groupinstall "Development Tools"
   dnf --enablerepo=devel install gpgme-devel rpmbuild


Put sources in place
================

.. code-block:: bash

   mkdir /root/rpmbuild/SOURCES
   cd /root/rpmbuild/SOURCES
   wget https://github.com/warewulf/warewulf/releases/download/v4.6.0rc2/warewulf-4.6.0rc2.tar.gz
   tar -xvzf warewulf-4.6.0rc2.tar.gz

Compile Warewulf and then build the RPM
================

.. code-block:: bash

   cd warewulf-4.6.0rc2
   make
   rpmbuild -ba warewulf.spec

Your RPMs are now present under /root/rpmbuild/RPMS
