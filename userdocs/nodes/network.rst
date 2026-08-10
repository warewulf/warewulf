.. _node-network:

==================
Network Interfaces
==================

By default, network configurations are applied to a "default" network interface.

.. code-block:: shell

  wwctl node set n001 \
    --netdev=eno1 \
    --hwaddr=00:00:00:00:00:01 \
    --ipaddr=10.0.2.1 \
    --netmask=255.255.255.0

Each cluster node can have multiple network interfaces, differentiated by
specifying  ``--netname``.

.. code-block:: shell

   wwctl node set n001 \
     --netname=infiniband \
     --type=infiniband \
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


.. _nettags:

Network Tags
============

Each network device can optionally have one or more key-value pair tags.

.. code-block:: shell

   wwctl node set n001 \
     --nettagadd="MYNETTAG=value"

.. _dns:

DNS
===

DNS of an interface (usually the default) can be configured using network tags.

- ``DNS1``: configures the first nameserver IP
- ``DNS2``: configures the second nameserver IP
- ``DNSSEARCH``: configures the DNS search domain

.. code-block:: shell

   wwctl node set n001 \
     --nettagadd "DNS1=10.0.0.1,DNS2=10.0.0.2,DNSSEARCH=my.domain.tld" \

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

   wwctl node set n001 \
     --netdev vlan42 \
     --ipaddr 10.0.42.1 \
     --netmask 255.255.252.0 \
     --netname iband \
     --type vlan \
     --nettagadd "vlan_id=42,parent_device=eth0"

.. _static_routes:

Static Routes
=============

The included Warewulf network overlays support the configuration of static routes
using a network tag of the form ``route<N>=<dest>,<gateway>``.

.. code-block:: shell

   wwctl node set n001 \
     --nettagadd "route1=192.168.2.0/24,192.168.1.254"

.. _hostname_suffix:

Hostname Suffix
===============

The ``host`` and ``hosts`` overlays support the customization of names suffixes in ``/etc/hosts``.

.. code-block:: shell

   wwctl node set n001 \
     --netname infiniband \
     --nettagadd "netsuffix=ib"
It will result in a content like the following.

.. code-block:: shell

   10.0.2.1 n001 n001-default n001-eno1
   10.0.3.1 n001-ib n001-ib1

.. _type:

Type
====

When using the NetworkManager overlay template the ``type`` attribute
determines how the connection is configured.  If not set, it
defaults to ``ethernet``. To correctly configure IPoIB with NetworkManager it
is required that ``type`` is set to ``infiniband``.
