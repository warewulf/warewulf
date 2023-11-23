=================
Warewulf Overlays
=================

So at this point, we have discussed how Warewulf is designed to
scalably provision and manage thousands of cluster nodes by utilizing
identical stateless boot images. And there-in lies a problem to
solve. If these boot images are completely identical, then how do we
configure things like hostnames? IP addresses? Or any other node
specific custom configurations?

While some of this can be managed by services like DHCP, and other
bits by configuration management, which can absolutely be done with
Warewulf and many people choose to do, these are heavy-weight
solutions to a simple problem to solve.

Warewulf solves this with overlays and uses overlays in different ways
through the provisioning process. A node or profile can configure an
overlay in two different ways:

* An overlay can be configured to run during boot as part of the
  ``wwinit`` process. These overlays are called **system overlay** or
  **wwinit overlays**.
* An overlay can be configured to run periodically while the system is
  running. These overlays are called **runtime overlays** or **generic
  overlays**.

The default profile has both a **wwinit** and a **runtime** overlay
configured.

Overlays are compiled for each compute node individually.

Defined Overlays
================

System or wwinit overlay
------------------------

This overlay contains all the nesscesary scripts to provision a
Warewulf node. It is available before the ``systemd`` or other init is
called and contains all configurations which are needed to bring up
the compute node. It is not updated during run time. Besides the
network configurations for

* wicked
* NetworkManager
* EL legacy network scripts

it also contains udev rules, which will set the interface name of the
first network device to ``eth0``.  Before the ``systemd`` init is
called, the overlay loops through the scripts in
``/wwinit/warwulf/init.d/*`` which will setup

* Ipmi
* wwclient
* selinux

Runtime Overlay or generic Overlay
----------------------------------

The runtime overlay is updated by the ``wwclient`` service on a
regular basis (by default, once per minute). In the standard
configuration it includes updates for ``/etc/passwd``, ``/etc/group``
and ``/etc/hosts``. Additionally the ``authorized_keys`` file of the
root user is updated.  It is recommended to use this overlay for
dynamic configuration files like ``slurm.conf``.  Once the system is
provisioned and booted, the ``wwclient`` program (which is provisioned
as part of the ``wwinit`` system overlay) will continuously update the
node with updates in the runtime overlay.

Host Overlay
------------

Configuration files used for the configuration of the Warewulf host /
server are stored in the **host** overlay. Unlike other overlays, it
*must* have the name ``host``. Existing files on the host are copied
to backup files with a ``wwbackup`` suffix at the first
run. (Subsequent use of the host overlay won't overwrite existing
``wwbackup`` files.)

The following services get configuration files via the host overlay:

* ssh keys are created with the scrips ``ssh_setup.sh`` and
  ``ssh_setup.csh``
* hosts entries are created by manipulating ``/etc/hosts`` with the
  template ``hosts.ww``
* nfs kernel server receives its exports from the template
  ``exports.ww``
* the dhcpd service is configured with ``dhcpd.conf.ww``

Combining Overlays
==================

When changing the overlays, it is recommended not to change them, but
to add the changed files to a new overlay and combine them in the
configuration. This is possible as the configuration fields for the
**wwinit** and **runtime** overlays are lists and can contain several
overlays.  As an example for this, we will overwrite the
``/etc/issue`` file from the **wwinit** overlay.  For this we will
create a new overlay called welcome and import the file ``/etc/issue``
from the host to it. This overlay is then combined with the existing
**wwinit** overlay.

.. code-block:: console

  # wwctl overlay create welcome
  # wwctl overlay mkdir welcome /etc
  # wwctl overlay import welcome /etc/issue
  # wwctl profile set default --wwinit=wwinit,welcome
  ? Are you sure you want to modify 1 profile(s)? [y/N] y
  # wwctl profile list default -a |grep welcome
  default              SystemOverlay      wwinit,welcome

Templates
=========

Templates allow you to create dynamic content such that the files
downloaded for each node will be customized for that node. Templates
allow you to insert everything from variables, to including files from
the control node, as well as conditional content and loops.

Warewulf uses the ``text/template`` engine to facilitate implementing
dynamic content in a simple and standardized manner.

All template files will end with the suffix of ``.ww``. That tells
Warewulf that when building a file, that it should parse that file as
a template. When it does that, the resulting file is static and can
have node customizations that are obtained from the node configuration
attributes.

.. note::

   When the file is persisted within the built overlay, the ``.ww``
   will be dropped, so ``/etc/hosts.ww`` will end up being
   ``/etc/hosts``.

Using Overlays
==============

Warewulf includes a command group for manipulating overlays (``wwctl
overlay``). With this you can add, edit, remove, change ownership,
permissions, etc.

..
  note::
  There is no possibility to delete files with an overlay!

Build
-----

.. code-block:: console

  wwctl overlay build [-H,--hosts|-N,--nodes|-o,--output directory|-O,--overlay-name] nodepattern

Without any arguments the command will interpret the templates for all
overlays for every compute node and also all the templates in the host
overlay. For every overlay of the compute nodes a gzip compressed cpio
archive is created. The range of the nodes can be restricted as the
last argument.  With the ``-H`` flag only the host overlay is
built. With the ``-N`` flag only compute node overlays are
built. Specific overlays can be selected with ``-O`` flag. For
debugging purposes the templates can be written to a directory given
via the ``-o`` flag.

By default Warewulf will build/update and cache overlays as needed
(configurable in the ``warewulf.conf``).

Chmod
-----

.. code-block:: console

  wwctl overlay chmod overlay-name filename mode

This subcommand changes the permissions of a single file within an
overlay. You can use any mode format supported by the chmod command.

Chown
-----

.. code-block:: console

  wwctl overlay chown overlay-name filename UID [GID]

With this command you can change the ownership of a file within a
given overlay to the user specified by UID. Optionally, it will also
change group ownership to GID.

Create
------

.. code-block:: console

  wwctl overlay create overlay-name

This command creates a new empty overlay with the given name.

Delete
------

.. code-block:: console

  wwctl overlay delete [-f,--force] overlay-name [File [File ...]]

Either the given overlay is deleted (must be empty or use the
``--force`` flag) or the specified file within the overlay is
deleted. With the ``--parents`` flag the directory of the deleted file
is also removed if no other file is in the directory.

Edit
----
.. code-block:: console

  wwctl overlay edit [--mode,-m MODE|--parents,-p] overlay-name file

Use this command to edit an existing or a new file in the given
overlay. If a the new file ends with a ``.ww`` suffix an example
template header is added to the file. With the ``--parents`` flag
necessary parent directories for a new file are created.

Import
------
.. code-block:: console

  wwctl overlay import [--mode,-m|--noupdate,-n] overlay-name file-name [new-file-name]

The given file is imported to the overlay. If no new-file-name is
given, the file will be placed in the overlay at the same path as on
the host. With the ``--noupdate`` flag you can block the rebuild of
the overlays.

List
----

.. code-block:: console

  wwctl overlay list [--all,-a|--long,-l] [overlay-name]

With this command all existing overlays and files in them can be
listed. Without any option only the overlay names and their number of
files are listed. With the ``--all`` switch also the every file is
shown. The ``--long`` option will also display the permissions, UID,
and GID of each file.

Show
----

.. code-block:: console

  wwctl overlay show [--quiet,-q|--render,-r nodename] overlay-name file

The content of the file for the given overlay is displayed with this
command. With the ``--render`` option a template is rendered as it
will be rendered for the given node. The node name is a mandatory
argument to the ``--render`` flag. Additional information for the file
can be suppressed with the ``--quiet`` option.
