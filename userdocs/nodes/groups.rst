===========
Node Groups
===========

Groups let you refer to a set of nodes by a single short name on the
``wwctl`` command line. Most ``wwctl`` subcommands that take a list of nodes
(``wwctl power``, ``wwctl ssh``, ``wwctl node list``, ``wwctl node set``,
``wwctl overlay build``, …) accept group references in addition to literal
node names and :ref:`hostlist <hostlist>` patterns. A group reference is the
name prefixed with ``@``.

.. code-block:: console

   # wwctl power reset @rack1
   # wwctl ssh @gpu uptime
   # wwctl overlay build @all

Declaring Group Membership
==========================

Group membership is declared on nodes and profiles via a ``groups:`` field;
the two sources are additive, so a node is in group ``G`` if it (or any
profile it inherits) lists ``G``.

.. code-block:: yaml

   nodeprofiles:
     gpu:
       groups:
         - gpu-nodes
     rack1:
       groups:
         - rack1
   nodes:
     n01:
       profiles:
         - gpu
         - rack1
     n02:
       profiles:
         - rack1
       groups:
         - admin
         - login

In the example above, ``@gpu-nodes`` resolves to every node that uses the
``gpu`` profile, ``@rack1`` resolves to ``n01`` and ``n02``, and ``@admin``
resolves to ``n02``.

The ``@all`` Built-in Group
===========================

The name ``all`` is reserved. ``@all`` always expands to every node defined
in ``nodes.conf``, even if no user-defined groups exist.

Opting a Node Out of ``@all``
=============================

Add the literal entry ``~all`` to a node's ``groups:`` field (or to a
profile the node inherits) to permanently exclude that node from ``@all``
expansion. This is the right knob for a head node, a quarantined node, or
anything else that should never be targeted by a bulk command issued
against ``@all``.

.. code-block:: yaml

   nodeprofiles:
     quarantine:
       groups:
         - ~all
   nodes:
     head01:
       groups:
         - ~all
     n02:
       profiles:
         - quarantine

``wwctl power reset @all`` will now skip both ``head01`` and ``n02``. The
exclusion does not affect explicit per-node commands (``wwctl power reset
head01`` still targets it directly), nor does it affect any user-defined
group the node is otherwise a member of.

There is one display gotcha: the merge step strips standalone ``~``-prefixed
entries, so ``wwctl node list head01 -a`` will not show ``~all`` under
``Groups``. To verify the opt-out took effect, run ``wwctl group list all``
and confirm the node is missing from the member list, or grep the raw
``nodes.conf``.

Negating an Inherited Group
===========================

A ``groups:`` entry prefixed with ``~`` removes a group that would
otherwise be inherited from a profile, the same way ``profiles:`` and
overlay lists support negation. For example, given a base profile that
places all nodes in the ``default`` group:

.. code-block:: yaml

   nodeprofiles:
     base:
       groups:
         - default
   nodes:
     n01:
       profiles:
         - base
     n02:
       profiles:
         - base
       groups:
         - ~default

``@default`` resolves to ``n01`` only; ``n02`` has been removed from the
group via negation.

Combining References on the Command Line
========================================

Plain node names, hostlists, and ``@group`` references can be mixed in a
single invocation. Duplicates are removed automatically.

.. code-block:: shell

   wwctl power reset n01 @rack2 login[01-02]

If a referenced group does not exist, Warewulf logs a warning and
contributes no nodes from that token; the rest of the command line still
resolves normally.

Inspecting Groups
=================

``wwctl group list`` enumerates every group referenced anywhere in the
configuration along with its members:

.. code-block:: console

   # wwctl group list
   GROUP  MEMBERS
   -----  -------
   admin  n02
   rack1  n01,n02,n03

Pass one or more names to filter the listing, including the built-in
``all``:

.. code-block:: console

   # wwctl group list rack1
   # wwctl group list all

To see which groups a *specific node* belongs to (and which source
contributed each), use ``wwctl node list <node> -a`` and look at the
``Groups`` row.
