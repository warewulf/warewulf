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

Syncing Local Users and Groups
=========

Warewulf now supports defining **node-specific local users and groups** through
the :ref:`resources <nodes-resources>` section of node definitions. These user and 
group entries are merged into the ``syncuser`` overlay and included in ``/etc/passwd`` and
``/etc/group`` on the target nodes.

This is useful for applications that require a node-local account (e.g.,
database, application, or storage service users) without having to maintain
centralized identity management.

.. note::

   Users created through this method do not have passwords set. They are intended for service accounts and non-interactive use.

Example:

.. code-block:: yaml

   nodes:
     n1:
       ...
        resources:
          localgroups:
            - gid: 1002
              members:
                - dbuser
                - dbuserbackup
              name: dbgroup
          localusers:
            - gid: 1001
              home: /
              name: dbuser
              shell: /bin/nologin
              uid: 1001
            - gid: 1005
              home: /
              name: dbuserbackup
              shell: /bin/nologin
              uid: 1005
        
When ``syncuser`` is executed on this node’s associated image, the ``dbuser``
and ``dbgroup`` entries will be added or updated accordingly.