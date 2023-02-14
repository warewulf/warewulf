=================
Warewulf Overlays
=================

So at this point, we have discussed how Warewulf is designed to scalably provision and manage thousands of cluster nodes by utilizing identical stateless boot images. And there-in lies a problem to solve. If these boot images are completely identical, then how do we configure things like hostnames? IP addresses? Or any other node specific custom configurations?

While some of this can be managed by services like DHCP, and other bits by configuration management, which can absolutely be done with Warewulf and many people choose to do, these are heavy-weight solutions to a simple problem to solve.

Warewulf solves this with overlays and uses overlays in different ways through the provisioning process. A node and profile can install this overlays at two times:

* Before boot, these overlays are called **system overlay** or **wwinit overlay**
* After boot into the running system which are the **runtime overlay** or **generic overlay**.

For both types preconfigured overlays are installed. Also for both types several overlays combined as list in the node/profile configuration. For conflicting files, the file of last defined one will be used.

Overlays are compiled for each compute node individually and should not contain static files.

Defined Overlays
================
Host Overlay
------------

In the host overlay the configuration files used for the configuration of the provision service are stored. In opposite the other overlays, it *must* have the name `host` and is stored under `/usr/share/warewulf/overlays/host/`.  Existing file on the host are copied to backup file with `wwbackup` suffix at the first run. Subsequent builds of the host overlay won't overwrite the `wwbackup` file.

Following services get configuration files via templates

* ssh for which are keys created with the scrips `ssh_setup.sh` and `ssh_setup.csh`
* hosts entries are created by manipulating `/etc/hosts` with the template `hosts.ww`
* nfs kernel server receives its exports from the template `exports.ww`
* the dhcpd service is configured with `dhcpd.conf.ww`

System or wwinit overlay
------------------------
This overlay contains all the nesscesary scripts for a warewulf installation. Its available before the `systemd` init is called and contains all configurations which are needed to bring up the compute node. It is not updated during run time.  Besides the network configurations for

* wicked
* NetworkManager
* EL legacy network scripts

it also contains udev rules, which will set the interface name of the first network device to `eth0`. 
Before the `systemd` init is called, the overlay loops through the scripts in `/wwinit/warwulf/init.d/*` which will setup
* Ipmi
* wwclient
* selinux

Runtime Overlay or generic Overlay
==================================

The runtime overlay is updated by the `wwclient` service on a regular base (every minute per default). In the standard configuration it includes updates for `/etc/passwd`, `/etc/group` and `/etc/hosts`. Additionally the `authorized_keys` file of the root user is updated.
It recommended to use this overlay for dynamic configuration files like `slurm.conf`.
Once the system is provisioned and booted, the ``wwclient`` program (which is provisioned as part of the ``wwinit`` system overlay) will continuously update the node with updates in the runtime overlay.

Templates
=========

Templates allow you to create dynamic content such that the files downloaded for each node will be customized for that node. Templates allow you to insert everything from variables, to including files from the control node, as well as conditional content and loops.

Warewulf uses the ``text/template`` engine to facilitate implementing dynamic content in a simple and standardized manner.

All template files will end with the suffix of ``.ww``. That tells Warewulf that when building a file, that it should parse that file as a template. When it does that, the resulting file is static and can have node customizations that are obtained from the node configuration attributes.

.. note::
   When the file is persisted within the built overlay, the ``.ww`` will be dropped, so ``/etc/hosts.ww`` will end up being ``/etc/hosts``.

Using Overlays
==============

Warewulf includes a command group for manipulating overlays (``wwctl overlay``). With this you can add, edit, remove, change ownership, permissions, etc.

Build
-----

.. code-block:: bash

  wwctl overlay build [-H,--hosts|-N,--nodes|-o,--output directory|-O,--overlay-name]  nodepattern

Without any arguments the command will interpret the templates for all overlays for every compute node and also all the templates in the host overlay. For every overlay of the compute nodes a gzip compressed cpio archive is created. The range of the nodes can be restricted as last argument.
With the `-H` flag only the host overlay is built, the `-N` flags restricts the build process to the compute nodes. Specific overlays can be selected with `-O` flag. For debugging purposes the templates can b written to a directory given via the `-o` flag.

By default Warewulf will build/update and cache overlays as needed (configurable in the `warewulf.conf`).

Chmod
-----

.. code-block:: bash

  wwctl overlay chmod overlay-name filename mode

This subcommand the permissions of a single file within an overlay.
You can use any mode format supported by the chmod command.

Chown
-----

.. code-block:: bash

  wwctl overlay chown overlay-name filename UID [GID]

With this command you can change the ownership of a file within a given overlay 
to the user specified by UID. Optionally, it will also change group ownership to GID

Create
------

.. code-block:: bash

  wwctl overlay create overlay-name

This command creates a new empty overlay with the given name.

Delete
------

.. code-block:: bash

  wwctl overlay delete [-f,--force] overlay-name [File [File ...]]

Either the given overlay is delete (must be empty or use the `--force flag`) or the file within the overlay is deleted. With the `--parents` flag also the directory of the delete file is removed, if no other file is in the directory.

Edit
----
.. code-block:: bash

  wwctl overlay edit [--mode,-m MODE|--parents,p]` overlay-name file

Use this command to edit an existing or a new template in the given overlay. If a the new file a `.ww` suffix an appropriate header is added to the file.  With the `--parents` flag necessary parent directories for a new file are created.

Import
------
.. code-block:: bash

  wwctl overlay import [--mode,-m|--noupdate,-n] overlay-name file-name [new-file-name]

The given file is imported to the overlay to the same place as it is on the host if no new file name is given. With the `--nodeupdate` flag you can  block the rebuild of the overlays

List
----

.. code-block:: bash

  wwctl overlay list [--all,a|--long,-l] [overlay-name`]

With this command all existing overlays and files in them can be listed. Without any option only the overlay names and their number of files are listed. With the `-all` switch also the every file is shown. The `--long` option will also display the permissions and UID,GID of a file.

Show
----

.. code-block:: bash

  wwctl overlay show [--quiet,-	q|--render,-r nodename] overlay-name file

The content of the file for the given overlay is displayed with this command. With the `--render` option a template is render as it will be rendered for the given node. The node name is a mandatory argument to the `--render` flag. Additional information for the file can be supressed vai the `--quiet` option.

