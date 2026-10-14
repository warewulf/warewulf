=========
Debugging
=========

Whether developing a new feature or fixing a bug, using the automated test suite
together with a debugger is a potent combination. This guide here can't
substitute for full documentation on a given debugger; but it might help you get
started debugging Warewulf.

Validating the code with vet
============================

The Warewulf ``Makefile`` includes a ``vet`` target which runs ``go vet`` on the
full codebase.

.. code-block:: shell

   make vet

Running the Full Test Suite
===========================

The Warewulf ``Makefile`` includes a ``test`` target which runs the full test
suite.

.. code-block:: shell

   make test

Individual test cases are particularly useful when coupled with a debugger. For
example, you can install delve as a regular user directly with Go.

.. code-block:: console

   $ go install github.com/go-delve/delve/cmd/dlv@latest

Visual Studio Code also includes a full-featured golang debugger that includes
testsuite integration.

Rendering Templates
===================

``wwctl overlay show --render <node> <overlay> <template>`` renders a template
for a node without building the overlay, and ``wwctl overlay info <overlay>
<template>`` lists the variables it uses.

To see which external commands ``wwctl`` runs (``cpio``, ``mksquashfs``, ...),
trace them:

.. code-block:: shell

   strace -f -e trace=execve wwctl overlay build <node>

Working in an Image
===================

Use ``wwctl image exec <image> -- <command>`` instead of ``chroot``; it sets up
``/proc``, ``/sys``, and ``/dev`` inside the image.

Debugging Node Boot
===================

Remove ``quiet`` from the node's kernel arguments and add ``rd.shell`` to get a
shell in the initramfs on failure. Set a serial console that matches the
hardware (``console=ttyS0``; ``console=ttyAMA0`` on aarch64 QEMU ``virt``).
