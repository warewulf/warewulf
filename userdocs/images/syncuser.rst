========
Syncuser
========

Warewulf can optionally synchronize UIDs and GIDs from the Warewulf server to an
image. This can be particularly useful when there is no central directory (e.g.,
an LDAP server).

.. note::
 
   Some system services (notably "munge") require a user to have the same UID across all nodes.

Combined with the "syncuser" overlay, Warewulf syncuser also supports defining
local users on the Warewulf server for synchronization to cluster nodes.

If there is mismatch between the server and the image, the import command will
generate a warning.

Syncuser may be invoked during image import, exec, shell, or build.

.. code-block:: console

   # wwctl image import --syncuser docker://ghcr.io/warewulf/warewulf-rockylinux:9 rockylinux-9
   # wwctl image exec --syncuser rockylinux-9 -- /usr/bin/echo "Hello, world!"
   # wwctl image shell --syncuser rockylinux-9
   # wwctl image build --syncuser rockylinux-9
   # wwctl image syncuser rockylinux-9

After syncuser, ``/etc/passwd`` and ``/etc/group`` in the image are updated, and
permissions on files belonging to these UIDs and GIDs are updated to match.
