==========
Nodegroups
==========

Nodegroups let you refer to a set of nodes by a single short name on the
``wwctl`` command line. Most ``wwctl`` subcommands that take a list of nodes
(``wwctl power``, ``wwctl ssh``, ``wwctl node list``, ``wwctl node set``,
``wwctl overlay build``, …) accept nodegroup references in addition to
literal node names and :ref:`hostlist <hostlist>` patterns. A nodegroup
reference is the name prefixed with ``@``.

.. code-block:: console

   # wwctl power reset @rack1
   # wwctl ssh @gpu uptime
   # wwctl overlay build @all

Nodegroup membership can be declared two different ways, and the two are
additive — a node is in nodegroup ``G`` if it appears in either source.

The Top-Level ``nodegroups:`` Stanza
====================================

A top-level ``nodegroups:`` stanza in ``nodes.conf`` maps each nodegroup
name to a list of node names. Hostlist range syntax is supported, so a
nodegroup can be described compactly:

.. code-block:: yaml

   nodegroups:
     rack1:
       - n[01-20]
     rack2:
       - n[21-40]
     login:
       - login01
       - login02

Per-Node and Per-Profile ``nodegroups:``
========================================

Individual nodes can also declare the nodegroups they belong to via a
``nodegroups:`` field. The same field is available on profiles, so a profile
such as ``gpu`` can automatically place every node that inherits it into a
``gpu`` nodegroup:

.. code-block:: yaml

   nodeprofiles:
     gpu:
       nodegroups:
         - gpu-nodes
   nodes:
     n01:
       profiles:
         - gpu
     n02:
       nodegroups:
         - admin
         - rack3

In the example above, ``@gpu-nodes`` resolves to every node that uses the
``gpu`` profile, ``@admin`` resolves to ``n02``, and the two mechanisms can
be mixed freely with the top-level ``nodegroups:`` stanza.

The ``@all`` Built-in Nodegroup
===============================

The name ``all`` is reserved. ``@all`` always expands to every node defined
in ``nodes.conf``, even if no user-defined nodegroups exist. Any user-defined
``nodegroups: all:`` stanza or per-node ``nodegroups: [all]`` entry is
harmless but redundant — ``@all`` is computed directly from the node list.

Opting a Node Out of ``@all``
=============================

Add the literal entry ``~all`` to a node's ``nodegroups:`` field (or to a
profile the node inherits) to permanently exclude that node from ``@all``
expansion. This is the right knob for a head node, a quarantined node, or
anything else that should never be targeted by a bulk command issued
against ``@all``.

.. code-block:: yaml

   nodeprofiles:
     quarantine:
       nodegroups:
         - ~all
   nodes:
     head01:
       nodegroups:
         - ~all
     n02:
       profiles:
         - quarantine

``wwctl power reset @all`` will now skip both ``head01`` and ``n02``. The
exclusion does not affect explicit per-node commands (``wwctl power reset
head01`` still targets it directly), nor does it affect any user-defined
nodegroup the node is otherwise a member of.

There is one display gotcha: the merge step strips standalone ``~``-prefixed
entries, so ``wwctl node list head01 -a`` will not show ``~all`` under
``NodeGroups``. To verify the opt-out took effect, run ``wwctl node group
list all`` and confirm the node is missing from the member list, or grep the
raw ``nodes.conf``.

Negating an Inherited Nodegroup
===============================

A ``nodegroups:`` entry prefixed with ``~`` removes a nodegroup that would
otherwise be inherited from a profile, the same way ``profiles:`` and
overlay lists support negation. For example, given a base profile that
places all nodes in the ``default`` nodegroup:

.. code-block:: yaml

   nodeprofiles:
     base:
       nodegroups:
         - default
   nodes:
     n01:
       profiles:
         - base
     n02:
       profiles:
         - base
       nodegroups:
         - ~default

``@default`` resolves to ``n01`` only; ``n02`` has been removed from the
nodegroup via negation.

Combining References on the Command Line
========================================

Plain node names, hostlists, and ``@nodegroup`` references can be mixed in
a single invocation. Duplicates are removed automatically.

.. code-block:: shell

   wwctl power reset n01 @rack2 login[01-02]

If a referenced nodegroup does not exist, Warewulf logs a warning and
contributes no nodes from that token; the rest of the command line still
resolves normally.

Inspecting Nodegroups
=====================

``wwctl node group list`` enumerates every nodegroup referenced anywhere in
the configuration along with its members (the union of all three sources):

.. code-block:: console

   # wwctl node group list
   NODEGROUP  MEMBERS
   ---------  -------
   admin      n02
   rack1      n01,n02,n03

Pass one or more names to filter the listing, including the built-in
``all``:

.. code-block:: console

   # wwctl node group list rack1
   # wwctl node group list all

To see which nodegroups a *specific node* belongs to (and which source
contributed each), use ``wwctl node list <node> -a`` and look at the
``NodeGroups`` row.
