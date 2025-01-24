========
Security
========

Historically, most HPC clusters utilize a security model that is "hard
on the exterior and soft and gushy on the interior". It is not that a
user has free roam once logged in, but rather we tend to rely on just
simple POSIX security models on the inside. For example, one of the
common practices is to completely disable SELinux on a new cluster
setup. Just kill it because it gets in the way.

For that reason, most critical HPC clusters leverage VPNs and/or
bastion hosts with multi-factor authentication (MFA) to help secure it
on the outside. But even with MFA and secure ssh connections through a
bastion host, it is still possible for malicious users to gain access
to these systems. Security being like layers of an onion is accurate,
but on an HPC system, those layers are predominately on the outside of
the cluster, not the inside.

Warewulf was written and designed from the ground up to go a bit
further. And while certain parallelization and high performance
library capabilities still require lowering the security threshold,
Warewulf strives to not be a blocker here.

SELinux
=======

The Warewulf server itself was developed with SELinux enabled in
"targeted" and "enforcing" mode and with the firewall active.

The provisioning process also fully supports booting SELinux-enabled
images, though nodes must be configured to use tmpfs for init. ("ramfs"
(often used by default) does not support extended file attributes.)

.. code-block:: bash

   wwctl profile set default --root tmpfs

.. note::

   Versions of Warewulf prior to v4.5.8 also required a kernel argument
   "rootfstype=ramfs" in order for wwinit to copy the node image to tmpfs; but
   this is no longer required.

Once that is done, you just need to enable SELinux in
``/etc/sysconfig/selinux`` and install the appropriate profiles into the
image. `An example`_ of such an image is available in the
warewulf-node-images repository.

.. _An example: https://github.com/warewulf/warewulf-node-images/tree/main/examples/rockylinux-9-selinux

Provisioning Security
=====================

Provisioning in generally is known to be rather "insecure" because
when a user lands on a compute node, there is generally nothing
stopping them from spoofing a provision request and downloading the
provisioned raw materials for inspection.

In Warewulf there are ways multiple to secure the provisioning process:

#. The provisioning connections and transfers are not secure due to
   not being able to manage a secure root of trust through a PXE
   process. The best way to secure the provisioning process is to
   enact a vLAN used specifically for provisioning. Warewulf supports
   this but you must consult your switch documentation and features to
   implement a default vLAN for provisioning and ensure that the
   runtime operating system is configured for a different tagged vLAN
   once booted.

#. Warewulf will leverage hardware "asset tags" which almost all
   vendors support. It is a configurable string that is configured in
   firmware and accessible only via root or physical access. During
   provisioning (as well as post provisioning via ``wwclient``)
   Warewulf, can use the asset tag as a secure token. If you have
   setup your hardware with an asset tag, you simply need to tell
   Warewulf what that asset tag is. When the asset tag is defined in
   Warewulf (``wwctl node set --assetkey "..."``), it will only
   provision and communicate with requests from that system matching
   that asset tag.

#. When the nodes are booted via `shim` and `grub` Secure Boot can be
   enabled. This means that the nodes only boot the kernel which is
   provided by the distributor and also custom complied modules can't
   be loaded.

Summary
=======

Warewulf does not limit the security posture of a cluster at all, and
perhaps it increases it as not all provisioners work with firewalls
and SELinux enabled and enforcing. But even with that, cluster
security is always up to the system manager and organizational
policies. Our job is just to ensure that we don't limit those policies
in any way.
