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

Many clusters have more than one private network. This is common for
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
* ``255.255.252.0``: Private network subnet mask (``10.0.0.0/22``)

Here is an example of how the cluster's address can be divided for a
255 node cluster:

* ``10.0.0.1 - 10.0.0.255``: Cluster infrastructure including this
  host, schedulers, file systems, routers, switches, etc.
* ``10.0.1.1 - 10.0.1.255``: DHCP range for booting nodes
* ``10.0.2.1 - 10.0.2.255``: Static node addresses
* ``10.0.3.1 - 10.0.3.255``: IPMI and/or out of band addresses for the
  compute nodes

Multiple networks
=================

It is possible to configure several networks not just for the nodes but also for the management of ``dhcpd`` and ``tftp``.
There are two ways to achive this:

1. Add the networks to the templates of ``dhcpd`` and/or the ``dnsmasq`` template directly.
2. Add the networks to a dummy node and change the templates of ``dhcp`` and ``dnsmasq`` accordingly.

As the first solution is trivial only the second way is described in the following lines.

As first the first step, add the dummy node.

.. code-block:: shell

   wwctl node add deliverynet

Now you will have to add the delivery networks to this node.

.. code-block:: shell

  wwctl node set \
    --ipaddr 10.0.20.250 \
    --netmask 255.255.255.0 \
    --netname deliver1 \
    --nettagadd network=10.0.20.0,dynstart=10.10.20.10,dynend=10.10.20.50 \
    deliverynet

  wwctl node set \
    --ipaddr 10.0.30.250 \
    --netmask 255.255.255.0 \
    --netname deliver2 \
    --nettagadd network=10.0.30.0,dynstart=10.10.30.10,dynend=10.10.30.50 \
    deliverynet

The ip address is used as the network address of host in the delivery network and an additional tags is used for definition of the network itself and the dynamic dhcp range.
You can check the result with ``wwctl node list``.

.. code-block:: console

  # wwctl node list -a deliverynet
  NODE         FIELD                             PROFILE  VALUE
  deliverynet  Id                                --       deliverynet
  deliverynet  Comment                           default  This profile is automatically included for each node
  deliverynet  ContainerName                     default  leap15.5
  deliverynet  Ipxe                              --       (default)
  deliverynet  RuntimeOverlay                    --       (generic)
  deliverynet  SystemOverlay                     --       (wwinit)
  deliverynet  Root                              --       (initramfs)
  deliverynet  Init                              --       (/sbin/init)
  deliverynet  Kernel.Args                       --       (quiet crashkernel=no vga=791 net.naming-scheme=v238)
  deliverynet  Profiles                          --       default
  deliverynet  PrimaryNetDev                     --       (deliver1)
  deliverynet  NetDevs[deliver2].Type            --       (ethernet)
  deliverynet  NetDevs[deliver2].OnBoot          --       (true)
  deliverynet  NetDevs[deliver2].Ipaddr          --       10.0.30.250
  deliverynet  NetDevs[deliver2].Netmask         --       255.255.255.0
  deliverynet  NetDevs[deliver2].Tags[dynend]    --       10.10.30.50
  deliverynet  NetDevs[deliver2].Tags[dynstart]  --       10.10.30.10
  deliverynet  NetDevs[deliver2].Tags[network]   --       10.0.30.0
  deliverynet  NetDevs[deliver1].Type            --       (ethernet)
  deliverynet  NetDevs[deliver1].OnBoot          --       (true)
  deliverynet  NetDevs[deliver1].Ipaddr          --       10.0.20.250
  deliverynet  NetDevs[deliver1].Netmask         --       255.255.255.0
  deliverynet  NetDevs[deliver1].Primary         --       (true)
  deliverynet  NetDevs[deliver1].Tags[network]   --       10.0.20.0
  deliverynet  NetDevs[deliver1].Tags[dynend]    --       10.10.20.50
  deliverynet  NetDevs[deliver1].Tags[dynstart]  --       10.10.20.10

Now the templates of ``dhcpd`` and/or ``dnsmasq`` must be modified.
You can edit the configuration files.

.. code-block:: shell

   wwctl overlay edit host etc/dhcpd.conf.ww
   wwctl overlay edit host etc/dnsmasq.d/ww4-hosts.ww

For the ``dhcp`` template you should add following lines

.. code-block::

   {{/* multiple networks */}}
   {{- range $node := $.AllNodes}}
   {{- if eq $node.Id.Get "deliverynet" }}
   {{- range $netname, $netdev := $node.NetDevs}}
   # network {{ $netname }}
   subnet {{$netdev.Tags.network.Get}} netmask {{$netdev.Netmask.Get}} {
       max-lease-time 120;
       range {{$netdev.Tags.dynstart.Get}} {{$netdev.Tags.dynend.Get}};
       next-server {{$netdev.Ipaddr.Get}};
   }
   {{- end }}
   {{- end }}
   {{- end }}

and for the ``dnsmasq`` the following lines should be added

.. code-block::

   {{/* multiple networks */}}
   {{- range $node := $.AllNodes}}
   {{- if eq $node.Id.Get "deliverynet" }}
   {{- range $netname, $netdev := $node.NetDevs}}
   # network {{ $netname }}
   dhcp-range={{$netdev.Tags.dynstart.Get}},{{$netdev.Tags.dynend.Get}},{{$netdev.Netmask.Get}},6h
   {{- end }}
   {{- end }}
   {{- end }}

Note that the ``{{- if eq $node.Id.Get "deliverynet" }}`` is used to identify the dummy host which carries the network information.
