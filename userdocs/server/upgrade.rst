==================
Upgrading Warewulf
==================

New versions of Warewulf might introduce changes to ``warewulf.conf`` and
``nodes.conf``. The ``wwctl upgrade`` command can help ease the transition
between versions.

.. note::

   ``wwctl upgrade`` will back up any files before it changes them (to
   ``<name>-old``) but it is good practice to back up your configuration
   manually.

.. code-block:: console

   # wwctl upgrade config
   # wwctl upgrade nodes --add-defaults --replace-overlays

Both upgrade commands support specifying ``--output-path=-`` to print the
upgraded configuration file to standard out for inspection before replacing the
configuration files.
