====================
Control Server Setup
====================

Operating System Installation
=============================

Warewulf has almost no predetermined or required configurations aside
from a base architecture networking layout. Install your Linux
distribution of choice as you would like, but do pay attention to the
cluster's private network configuration.

Network
=======

A clustered resource depends on a private management network. This
network can be either persistent (it is always "up" even after
provisioning) to temporary which might only be used for provisioning
and/or out of band system control and management e.g. IPMI).

It is important for this management network to be private to the
compute resource because Warewulf requires network services on that
network which may conflict with services on the production/public
network (e.g. DHCP). It is also important from a security perspective
as the management network for typical HPC systems have an implied
trust level associated with it and generally there is no firewalling
or network monitoring occurring on these networks.

Usually, the control node is "dual homed" which means it has at least
two interface cards, one connected to the private cluster network and
one dedicated to the public network (as the figure above
demonstrates).

.. note::

   It is possible to omit the public network interface with a reverse
   NAT. Warewulf can operate in this configuration but it extends
   beyond the scope of this documentation.

Many clusters have more then one private network. This is common for
performance critical HPC clusters that implement a high speed and low
latency network like InfiniBand. In this case, this network is used
for high speed data transfers for inter-process communication between
compute nodes and file system IO.

Warewulf will need to be configured to use the private cluster
management network. Warewulf will use this network for booting the
nodes over PXE. There are three network protocols used to accomplish
this DHCP/BOOT, TFTP, and HTTP on port ``9873``. Warewulf will use the
operating system's provided version of DHCP (ISC-DHCP) and TFTP for
the PXE bootstrap to iPXE, and then iPXE will use Warewulf's internal
HTTP services to transfer the larger files for provisioning.

Addressing
==========

The addressing scheme of your private cluster network is 100% up to
the system integrator, but for large clusters, many organizations like
to organize the address allocations. Below is a recommended IP
addressing scheme which we will use for the rest of this document.

* ``10.0.0.1``: Private network address IP
* ``255.255.252.0``: Private network subnet mask

Here is an example of how the cluster's address can be divided for a
255 node cluster:

* ``10.0.0.1 - 10.0.0.255``: Cluster infrastructure including this
  host, schedulers, file systems, routers, switches, etc.
* ``10.0.1.1 - 10.0.1.255``: DHCP range for booting nodes
* ``10.0.2.1 - 10.0.2.255``: Static node addresses
* ``10.0.3.1 - 10.0.3.255``: IPMI and/or out of band addresses for the
  compute nodes
