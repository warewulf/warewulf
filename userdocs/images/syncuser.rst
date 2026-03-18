========
Syncuser
========

Warewulf's syncuser feature has two distinct parts that work together:

1. The **syncuser command** synchronizes UIDs and GIDs from the Warewulf server
   into an OS image, updating ``/etc/passwd``, ``/etc/group``, and file
   ownerships within the image.

2. The **syncuser overlay** merges users and groups from both the OS image and
   the Warewulf server into the provisioned node's ``/etc/passwd`` and
   ``/etc/group`` at runtime.

This is particularly useful when there is no central directory (e.g., an LDAP
server). Together, these two parts ensure that UIDs and GIDs are consistent
across the server and all cluster nodes.

.. note::

   Some system services (notably "munge") require a user to have the same UID
   across all nodes.

Synchronizing UIDs/GIDs in an OS image
--------------------------------------

The syncuser command updates an OS image so that any user or group present on
the Warewulf server has the same UID/GID inside the image. Users and groups
that exist only in the image are preserved unless a UID/GID collision is
detected. File ownerships within the image are also updated to match.

If there is a mismatch between the server and the image, the import command
will generate a warning.

Syncuser may be invoked during image import, exec, shell, or build:

.. code-block:: console

   # wwctl image import --syncuser docker://ghcr.io/warewulf/warewulf-rockylinux:9 rockylinux-9
   # wwctl image exec --syncuser rockylinux-9 -- /usr/bin/echo "Hello, world!"
   # wwctl image shell --syncuser rockylinux-9
   # wwctl image build --syncuser rockylinux-9
   # wwctl image syncuser rockylinux-9

After syncuser, ``/etc/passwd`` and ``/etc/group`` in the image are updated,
and permissions on files belonging to these UIDs and GIDs are updated to match.

The syncuser overlay
--------------------

The syncuser overlay runs at provisioning time and merges ``/etc/passwd`` and
``/etc/group`` from both the OS image and the Warewulf server. This makes users
defined on the server (but not originally in the image) available on provisioned
nodes.

For the overlay to work correctly, the image should also have been prepared with
the syncuser command (see above) so that UID/GID values are consistent.

See :ref:`Syncuser` in the overlays documentation for more detail.
