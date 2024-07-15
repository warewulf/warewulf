============================
Controlling Warewulf (wwctl)
============================

Warewulf's command-line interface is based primarily around the
``wwctl`` command. This command has sub-commands for each major
component of Warewulf's functionality.

* ``configure`` configures system services that Warewulf depends on
* ``container`` configures containers (node images)
* ``kernel`` configures override kernels
* ``node`` manages nodes in the cluster
* ``profiles`` defines configuration which can be applied to multiple
  nodes
* ``overlays`` manages overlays
* ``power`` turns nodes on and off
* ``ssh`` provides basic parallel ssh functionality

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
