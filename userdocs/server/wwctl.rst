====================
Controlling Warewulf
====================

Warewulf's command-line interface is based primarily around the ``wwctl``
command. This command has sub-commands for each major component of Warewulf's
functionality.

* ``configure``: configures the Warewulf server and its external services
* ``node``: manages nodes in the cluster
* ``profiles``: defines common sets of node configuration which can be applied
  to multiple nodes
* ``image``: configures (node) images
* ``overlays``: manages overlays

``wwctl`` also provides additional helpers for interacting with cluster nodes
over SSH and IPMI.

* ``power``: turns nodes on and off
* ``ssh``: provides basic parallel ssh functionality

All of these subcommands (and their respective sub-subcommands) have
built-in help with either ``wwctl help`` or ``--help``.

Hostlists
=========

Many of the commands (e.g., ``wwctl node list`` support a "hostlist"
syntax for referring to multiple nodes at once. Hostlist expressions
support both ranges and comma-separated numerical lists.

For example:

* ``node[1-2]`` expands to ``node1 node2``
* ``node[1,3]`` expands to ``node1 node3``
* ``node[1,5-6]`` expands to ``node1 node5 node6``

Node status
===========
During the whole provisioning process of your nodes, you can check their status
through the following command :

.. code-block:: console

   # wwctl node status
   NODENAME             STAGE                SENT                      LASTSEEN (s)
   ================================================================================
   n1                   RUNTIME_OVERLAY      __RUNTIME__.img.gz        16

For each node, there are 4 different stages:

* **IPXE**
* **KERNEL**
* **SYSTEM_OVERLAY**
* **RUNTIME_OVERLAY**

You can use the ``wwctl node status`` to check communication between the
Warewulf server (``warewulfd``) and the Warewulf client (``wwclient``).
