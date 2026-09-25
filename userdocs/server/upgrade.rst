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

Per-service host overlays
=========================

Warewulf no longer applies a single ``host`` overlay implicitly. Each service
configured by ``wwctl configure`` applies only the host overlays listed in its
``overlays`` key in ``warewulf.conf``:

.. code-block:: yaml

   dhcp:
     overlays: dhcpd
   tftp:
     overlays: tftproot
   nfs:
     overlays: nfsd
   ssh:
     overlays: ssh.wwctl
   hostfile:
     overlays: hosts

There are no defaults for these keys. A service with no ``overlays`` listed is
skipped with a warning, and its configuration files (for example
``dhcpd.conf``, ``/etc/exports``, ``/etc/hosts``, or ``grub.cfg``) are no
longer updated.

Package upgrades do not replace an existing ``warewulf.conf``, so run ``wwctl
upgrade config`` after upgrading to add the recommended ``overlays`` for each
service, then re-run ``wwctl configure --all``.

If a site ``host`` overlay is still present, ``wwctl upgrade config`` also appends
``host`` to each service's ``overlays`` so that its customizations continue to
be applied. Move those customizations into the overlay for the relevant service
and remove the ``host`` entries. See :ref:`host-overlays`.
