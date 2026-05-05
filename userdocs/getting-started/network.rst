================
Network Planning
================

A clustered resource depends on a cluster network. This network can be either
persistent (it is always "up" even after provisioning) or temporary, only used
for provisioning and/or out of band system control and management (e.g., IPMI).

The cluster network must be dedicated to the cluster because Warewulf uses
network services (particularly DHCP) which may conflict with services on another
mixed-use network. A dedicated cluster network is also important for security,
as the cluster network often has an implicit level of trust associated with it.

The Warewulf server is often "dual homed," meaning that it has separate network
interfaces connected to each of the cluster network and an external network. But
it is also possible for the cluster network to be routable from other, more
general-purpose networks.

Many clusters have more than one internal network. This is common for
performance critical HPC clusters that implement a high speed and low latency
network like InfiniBand. In this case, this network is used for high speed data
transfers for inter-process communication between compute nodes and file system
IO.

Warewulf will need to be configured to use the private cluster management
network. Warewulf will use this network for booting the nodes over PXE. There
are three network protocols used to accomplish this DHCP/BOOT, TFTP, and HTTP on
port ``9873``. Warewulf will use the operating system's provided version of DHCP
(ISC-DHCP) and TFTP for the PXE bootstrap to iPXE, and then iPXE will use
Warewulf's internal HTTP services to transfer the larger files for provisioning.

Addressing
==========

The addressing scheme of your private cluster network is 100% up to the system
integrator, but for large clusters, many organizations like to organize the
address allocations. Below is a recommended IP addressing scheme which we will
use for the rest of this document.

* ``10.0.0.1``: Private network address IP
* ``255.255.252.0``: Private network subnet mask (``10.0.0.0/22``)

Here is an example of how the cluster's address can be divided for a 255 node
cluster:

* ``10.0.0.1 - 10.0.0.255``: Cluster infrastructure including this
  host, schedulers, file systems, routers, switches, etc.
* ``10.0.1.1 - 10.0.1.255``: DHCP range for booting nodes
* ``10.0.2.1 - 10.0.2.255``: Static node addresses
* ``10.0.3.1 - 10.0.3.255``: IPMI and/or out of band addresses for the
  compute nodes
