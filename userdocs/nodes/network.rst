.. _node-network:

==================
Network Interfaces
==================

By default, network configurations are applied to a "default" network interface.

.. code-block:: shell

  wwctl node set n1 \
    --netdev=eno1 \
    --hwaddr=00:00:00:00:00:01 \
    --ipaddr=10.0.2.1 \
    --netmask=255.255.255.0

Each cluster node can have multiple network interfaces, differentiated by
specifying  ``--netname``.

.. code-block:: shell

   wwctl node set n1 \
     --netname=infiniband \
     --netdev=ib1 \
     --ipaddr=10.0.3.1 \
     --netmask=255.255.255.0

.. warning::

   Due to the way network interface names are assigned by the Linux kernel, and
   later reassigned by udev and systemd, the use of ``eth0``, ``eth1``, etc. as
   interface is strongly discouraged. We recommend the use of the original
   predictable names assigned to the interfaces (e.g., ``eno1``), as otherwise
   an interface may fail to be named correct if its desired name conflicts with
   the kernel-assigned name of another interface during the boot process.

.. _bonding:

Bonding
=======

Support for bonded / link aggregation network interfaces depends on the network overlay being used.

The ``ifcfg`` and ``NetworkManager`` overlays can configure a network bond like this:

.. code-block:: yaml

   network devices:
     bond0:
       type: Bond
       device: bond0
       ipaddr: 192.168.3.100
       netmask: 255.255.255.0
     en1:
       device: en1
       hwaddr: e6:92:39:49:7b:03
       tags:
         master: bond0
     en2:
       device: en2
       hwaddr: 9a:77:29:73:14:f1
       tags:
         master: bond0

.. _vlan:

VLAN
====

You can set the type also to ``vlan``.

Some network configuration systems use the network device name
(e.g., of the form ``eno1.100``)
to configure VLANs.
Other network systems need additional network tags:

- ``vlan_id``: configures the VLAN ID of the interface
- ``parent_device``: configures which physical interface to use

.. code-block:: shell

   wwctl node set \
     --netdev vlan42 \
     --ipaddr 10.0.42.1 \
     --netmask 255.255.252.0 \
     --netname iband \
     --type vlan \
     --nettagadd "vlan_id=42,parent_device=eth0" \
     n001

.. _static_routes:

Static Routes
=============

The included Warewulf network overlays support the configuration of static routes
using a network tag of the form ``route<N>=<dest>,<gateway>``.

.. code-block:: shell

   wwctl node set n001 \
     --nettagadd "route1=192.168.2.0/24,192.168.1.254"