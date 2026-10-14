===============
Code Guidelines
===============

Warewulf is written in Go. This page describes how the source is laid out and
the conventions that keep it consistent. Read it before changing Go code, and
follow the existing precedent where this page is silent.

Source Layout
=============

* ``cmd/``: ``main`` packages for ``wwctl`` and ``wwclient``. They only call
  into ``internal/app/``.
* ``internal/app/wwctl/``: one package per ``wwctl`` subcommand, nested like
  the command line (e.g. ``internal/app/wwctl/node/set/``).
* ``internal/app/wwclient/``: the node-side client.
* ``internal/pkg/``: shared packages. The ones most changes touch:

  * ``node``: the ``nodes.conf`` data model (nodes and profiles).
  * ``config``: the ``warewulf.conf`` data model.
  * ``overlay``: overlay building, template rendering, and the template
    function map.
  * ``image``: node image import and build.
  * ``warewulfd``: the provisioning daemon and its REST API.
  * ``wwlog``: logging.
  * ``util``: general helpers (files, cpio, YAML, ``BoolP``).
  * ``hostlist``: node-range expansion (``n[01-10]``).
  * ``testenv``: the test fixture (see :ref:`writing-tests`).

* ``overlays/<name>/rootfs/``: the distribution overlays; their tests are in
  ``overlays/<name>/internal/``.
* ``etc/``, ``lib/``, ``include/``, ``dracut/``: default configuration, BMC
  templates, systemd and other installed files, and the ``wwinit`` dracut
  module.

wwctl Subcommands
=================

Each subcommand package has three files:

* ``root.go``: ``GetCommand()`` builds the ``cobra.Command``, registers its
  flags and completions, and sets ``RunE: CobraRunE(&vars)``.
* ``main.go``: ``CobraRunE`` does the work.
* ``main_test.go``: tests that run ``GetCommand()`` with arguments.

Use the shared helpers in ``internal/app/wwctl/completions`` and
``internal/app/wwctl/flags`` for completions and common flags. Return errors
from ``RunE``, wrapped with context (``fmt.Errorf("...: %w", err)``), rather
than exiting.

Node and Profile Fields
=======================

``nodes.conf`` fields are struct fields in ``internal/pkg/node``. Their struct
tags drive the rest: ``yaml`` and ``json`` name the field, ``lopt`` and
``sopt`` name the ``wwctl`` long and short flags, and ``comment`` is the help
text. ``wwctl node set`` and ``wwctl profile set`` generate their flags from
these tags, so a new field needs only the tags, not hand-written flag code.
Document a new field in the user guide and in ``docs/man/man5/nodes.conf.5``.

``warewulf.conf`` fields are in ``internal/pkg/config``, with defaults set by
``default`` struct tags. Document a new field in
``userdocs/server/configuration.rst`` and ``docs/man/man5/warewulf.conf.5``.

If a change alters the on-disk format of either file, check whether
``internal/pkg/upgrade`` (``wwctl upgrade``) needs a matching step so existing
configurations still read.

Logging and Output
==================

Use ``internal/pkg/wwlog`` rather than ``fmt`` or ``log``. Choose the level by
audience: ``Info`` and ``Output`` are what the user asked for and go to
stdout; ``Verbose`` (``wwctl -v``) and ``Debug`` (``wwctl -d``) are detail;
``Warn`` and ``Error`` report problems. Everything except ``Info`` and
``Output`` goes to stderr. The ``*Exc`` variants take an ``error``.

Existing Helpers
================

Before writing a helper, check ``internal/pkg/util`` and the standard library
(``slices``, ``maps``, ``path/filepath``). Expand node arguments with
``hostlist.Expand``. When behavior depends on an external tool that differs
between distributions (``cpio``, ``dracut``), detect the feature rather than
comparing versions; ``util.CpioCreate`` is an example.

A new overlay template function goes in ``buildTemplateFuncMap`` in
``internal/pkg/overlay/overlay.go``, with a section in
:ref:`templates` and a case in
``internal/pkg/overlay/overlay_test.go``.

Generated Files
===============

Some files are generated. Edit their sources and let ``make`` regenerate them:

* ``internal/pkg/config/buildconfig.go``, ``warewulf.spec`` and
  ``include/systemd/warewulfd.service`` come from their ``.in`` files
  (``make config``).
* ``userdocs/reference/`` and ``docs/man/man1/`` come from ``wwctl`` itself
  (``make docs``).
* ``LICENSE_DEPENDENCIES.md`` comes from
  ``scripts/update-license-dependencies.sh``; CI checks that it is current.

``make`` targets that need the generated files create them. Run ``make config``
first if you call ``go`` directly.

Checks
======

CI (``.github/workflows/testsuite.yml``) runs the same targets you can run
locally:

.. code-block:: shell

   make fmt
   make lint vet staticcheck deadcode shellcheck
   make test

``deadcode`` fails on any unused function, so remove helpers you stop using.
``lint``, ``deadcode`` and ``shellcheck`` download their tools into ``.tools/``
on first use, and builds vendor modules first; set ``OFFLINE_BUILD=1`` to skip
both when there is no network access.
