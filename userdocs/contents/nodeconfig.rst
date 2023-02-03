==================
Node Configuration
==================

The Node Configuration DB
=========================

As mentioned in the [Configuration](configuration) section, node
configs are persisted to the ``nodes.conf`` YAML file, but generally it
is best not to edit this file directly (however that is supported, it
is just prone to errors).

This method of using a YAML configuration file as a backend datastore
is both scalable and very lightweight. We've tested this out to over
10,000 node entries which yielded update latencies under 1 second,
which we felt was both tolerable and advantageous.

Adding a New Node
=================

Creating a new node is as simple as running the following command:

.. code-block:: bash

   $ sudo wwctl node add n0000
   Added node: n0000

Node Names
----------

For small clusters, you can use simple names (e.g. ``n0000``); but for
larger, more complicated clusters that are comprised of multiple
clusters and roles it is highly recommended to use node names that
include a cluster descriptor. In Warewulf, this is generally done by
using a domain name (e.g. ``n0000.cluster01``). Warewulf will
automatically assume that the domain is the equivalent of the cluster
name.

This also means that you can address groups of nodes by the cluster
descriptor with globs. For example, you are able to refer to all nodes
in "cluster01" with the following string: ``*.cluster01`` which is
valuable for other ``wwctl`` commands.

Listing Nodes
=============

Once you have configured one or more nodes, you can list them and
their attributes as follows:

.. code-block:: bash

   $ sudo wwctl node list
   NODE NAME              PROFILES                   NETWORK
   ================================================================================
   n0000                  default

You can also see the node's full attribute list by specifying the ``-a``
option (all):

.. code-block:: bash

   $ sudo wwctl node list -a
   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   n0000                Id                 --           n0000
   n0000                Comment            default      This profile is automatically included for each node
   n0000                Cluster            --           --
   n0000                Profiles           --           default
   n0000                Discoverable       --           false
   n0000                Container          --           --
   n0000                KernelOverride     --           --
   n0000                KernelArgs         --           (quiet crashkernel=no vga=791 rootfstype=rootfs)
   n0000                RuntimeOverlay     --           (default)
   n0000                SystemOverlay      --           (default)
   n0000                Ipxe               --           (default)
   n0000                Init               --           (/sbin/init)
   n0000                Root               --           (initramfs)
   n0000                IpmiIpaddr         --           --
   n0000                IpmiNetmask        --           --
   n0000                IpmiPort           --           --
   n0000                IpmiGateway        --           --
   n0000                IpmiUserName       --           --
   n0000                IpmiInterface      --           --
   n0000                IpmiWrite          --           --

.. note::
   The attribute values in parenthesis are default values and can
   be overridden in the next section, granted, the default values are
   generally usable.

Setting Node Attributes
=======================

In the above output we can see that there is no kernel or container
defined for this node. To provision a node, the minimum requirements
are a kernel and container, and for that node to be useful, we will
also need to configure the network so the nodes are reachable after
they boot.

Node configurations are set using the ``wwctl node set`` command. To see
a list of all configuration attributes, use the command ``wwctl node
set --help``.

Configuring the Node's Container Image
======================================

.. code-block:: bash

   $ sudo wwctl node set --container rocky-8 n0000
   Are you sure you want to modify 1 nodes(s): y

And you can check that the container name is set for ``n0000``:

.. code-block:: bash

   $ sudo wwctl node list -a  n0000 | grep Container
   n0000                Container          --           rocky-8

Configuring the Node's Kernel
-----------------------------

While the recommended method for assigning a kernel in 4.3 and beyond
is to include it in the container / node image, a kernel can still be
specified as an override at the node or profile.

.. code-block:: bash

   $ sudo wwctl node set --kerneloverride $(uname -r) n0000
   Are you sure you want to modify 1 nodes(s): y

   $ sudo wwctl node list -a n0000 | grep KernelOverride
   n0000                KernelOverride     --           4.18.0-305.3.1.el8_4.x86_64

Configuring the Node's Network
------------------------------

To configure the network, we have to pick a network device name and
provide the network information as follows:

.. code-block:: bash

   $ sudo wwctl node set --netdev eth0 --hwaddr 11:22:33:44:55:66 --ipaddr 10.0.2.1 --netmask 255.255.252.0 n0000
   Are you sure you want to modify 1 nodes(s): y

You can now see that the node contains configuration attributes for
container, kernel, and network:

.. code-block:: bash

   $ sudo wwctl node list -a n0000
   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   n0000                Id                 --           n0000
   n0000                Comment            default      This profile is automatically included for each node
   n0000                Cluster            --           --
   n0000                Profiles           --           default
   n0000                Discoverable       --           false
   n0000                Container          --           rocky-8
   n0000                Kernel             --           4.18.0-305.3.1.el8_4.x86_64
   n0000                KernelArgs         --           (quiet crashkernel=no vga=791 rootfstype=rootfs)
   n0000                RuntimeOverlay     --           (default)
   n0000                SystemOverlay      --           (default)
   n0000                Ipxe               --           (default)
   n0000                Init               --           (/sbin/init)
   n0000                Root               --           (initramfs)
   n0000                IpmiIpaddr         --           --
   n0000                IpmiNetmask        --           --
   n0000                IpmiPort           --           --
   n0000                IpmiGateway        --           --
   n0000                IpmiUserName       --           --
   n0000                IpmiInterface      --           --
   n0000                default:DEVICE     --           eth0
   n0000                default:HWADDR     --           11:22:33:44:55:66
   n0000                default:IPADDR     --           10.0.2.1
   n0000                default:NETMASK    --           255.255.252.0
   n0000                default:GATEWAY    --           --
   n0000                default:TYPE       --           --
   n0000                default:DEFAULT    --           false

Un-setting Node Attributes
==========================

If you wish to ``unset`` a particular value, set the value to
``UNDEF``. For example:

.. code-block:: bash

   $ sudo wwctl node set --cluster cluster01 n0000
   Are you sure you want to modify 1 nodes(s): y

   $ sudo wwctl node list -a n0000 | grep Cluster
   n0000                Cluster            --           cluster01

And to unset this configuration attribute:

.. code-block:: bash

   $ sudo wwctl node set --cluster UNDEF n0000
   Are you sure you want to modify 1 nodes(s): y

   $ sudo wwctl node list -a n0000 | grep Cluster
   n0000                Cluster            --           --