===========
Node Groups
===========

Groups let you refer to a set of nodes by a single short name on the
``wwctl`` command line. Most ``wwctl`` subcommands that take a list of nodes
(``wwctl power``, ``wwctl ssh``, ``wwctl overlay build``, …) accept group
references in addition to literal node names and :ref:`hostlist <hostlist>`
patterns. A group reference is the name prefixed with ``@``.

.. code-block:: console

   # wwctl power reset @rack1
   # wwctl ssh @gpu uptime
   # wwctl overlay build @chemistry

Declaring Group Membership
==========================

Group membership is declared on nodes and profiles via a ``groups:`` field;
the two sources are additive, so a node is in group ``foo`` if it (or any
profile it inherits) lists ``foo``.

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
     head01:
       groups:
         - admin
         - ~all

In the example above, ``@gpu-nodes`` resolves to every node that uses the
``gpu`` profile, ``@rack1`` resolves to ``n01`` and ``n02``, and ``@admin``
resolves to ``n02``. ``head01`` is included in ``@admin`` but excluded from
``@all``.

The ``@all`` Built-in Group
===========================

The name ``all`` is reserved. ``@all`` always expands to every node defined
in ``nodes.conf``, even if no user-defined groups exist. Nodes can be excluded
from ``all`` by negating it via node or profile config.

Excluding from a group
======================

A ``groups:`` entry prefixed with ``~`` removes a group that would
otherwise be inherited from a profile, the same way ``profiles:`` and
overlay lists support negation. For example, given a ``lab-course`` profile
that places all nodes in the ``interactive`` group:

.. code-block:: yaml

   nodeprofiles:
     lab-course:
       groups:
         - interactive
   nodes:
     n01:
       profiles:
         - lab-course
     n02:
       profiles:
         - lab-course
       groups:
         - ~interactive

``@interactive`` resolves to ``n01`` only; ``n02`` has been removed from the
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

Pass ``--noheader`` / ``-n`` together with one or more group names to get
just the membership as a single comma-separated, deduped list — useful for
feeding the membership into non-``wwctl`` tools:

.. code-block:: console

   # wwctl group list -n lab-course
   n01,n02,n03
   # scontrol create reservation [...] nodes=$(wwctl group list -n lab-course)

Inspecting Nodes
================

To see which groups a *specific node* belongs to (and which source
contributed each), use ``wwctl node list <node> -a`` and look at the
``Groups`` row.
