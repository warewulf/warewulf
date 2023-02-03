=================
Warewulf Overlays
=================

So at this point, we have discussed how Warewulf is designed to scalably provision and manage thousands of cluster nodes by utilizing identical stateless boot images. And there-in lies a problem to solve. If these boot images are completely identical, then how do we configure things like hostnames? IP addresses? Or any other node specific custom configurations?

While some of this can be managed by services like DHCP, and other bits by configuration management, which can absolutely be done with Warewulf and many people choose to do, these are heavy-weight solutions to a simple problem to solve.

Warewulf solves this with overlays and uses overlays in different ways through the provisioning process. Two of these overlays are exposed to the users, the **system overlay** and the **runtime overlay**.

.. note::
   Another overlay that isn't directly exposed is the **kmods overlay** which contains all of the kernel modules to match the configured kernel. Because this overlay is used "behind the scenes" it is outside the scope of this document.

System Overlay
==============

The System Overlay is used by the core of Warewulf to setup the environment on the node necessary for provisioning. The default system overlay is called ``wwinit``. Generally speaking, it will not be necessary to make changes to this overlay, but it is possible to change or configure this overlay to meet site specific needs if necessary.

Runtime Overlay
===============

The Runtime Overlay is the overlay that is responsible for most of the typical system administration configurations. Here you will make changes necessary to support your operating system as well as application configurations.

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

The general syntax is as follows:

.. code-block:: bash

   wwctl overlay [action] [overlay name] ...

* **action**: the overlay subcommand you are invoking
* **overlay name**: the name of the overlay in question within a given type
* **...**: additional arguments are action specific

By default there is one overlay in each of the system and runtime overlay types. Both overlays are called "default". To say it differently, there are two default overlays, one is a system overlay and one is a runtime overlay.

Viewing the Files Within an Overlay
===================================

Overlays can be viewed with the command ``wwctl overlay list``. You can see the files within an overlay by adding the ``-a`` or ``-l`` options as follows:

.. code-block:: bash

   $ sudo wwctl overlay list -l generic
   PERM MODE    UID GID   OVERLAY    FILE PATH
   -rwxr-xr-x     0 0     generic            /etc/
   -rw-r--r--     0 0     generic            /etc/group.ww
   -rw-r--r--     0 0     generic            /etc/hosts.ww
   -rw-r--r--     0 0     generic            /etc/passwd.ww
   -rwxr-xr-x     0 0     generic            /root/
   -rwxr-xr-x     0 0     generic            /root/.ssh/
   -rw-r--r--     0 0     generic            /root/.ssh/authorized_keys.ww

Creating a New File Within an Overlay
=====================================

Just like any file on the system, you can create and edit a file at the same time. So to do that, you simple ``edit`` a new file as follows:

.. code-block:: bash

   $ sudo wwctl overlay edit [overlay name] [file path]

For example:

.. code-block:: bash

   $ sudo wwctl overlay edit generic /etc/testfile

and you can validate that the file is there with the ``list`` command:

.. code-block:: bash

   $ sudo wwctl overlay list generic -l
   PERM MODE    UID GID   RUNTIME-OVERLAY    FILE PATH
   -rwxr-xr-x     0 0     generic            /etc/
   -rw-r--r--     0 0     generic            /etc/group.ww
   -rw-r--r--     0 0     generic            /etc/hosts.ww
   -rw-r--r--     0 0     generic            /etc/passwd.ww
   -rwxr-xr-x     0 0     generic            /etc/testfile
   -rwxr-xr-x     0 0     generic            /root/
   -rwxr-xr-x     0 0     generic            /root/.ssh/
   -rw-r--r--     0 0     generic            /root/.ssh/authorized_keys.ww

.. note::
   To create a template file, simply name the file with the suffix ``.ww``. This suffix will tell Warewulf that the file should be parsed by the templating engine and written into the overlay with the suffix stripped off.

Building Overlays
=================

By default Warewulf will build/update and cache overlays as needed (configurable in the ``warewulf.conf``).

You can however build overlays by hand, and in some cases this will be advantageous (like if you are freshly booting thousands of compute nodes in parallel). The command to do that is:

.. code-block:: bash

   # wwctl overlay build n00[00-10]
   Building overlays for n0000: [wwinit, generic]
   Building overlays for n0001: [wwinit, generic]
   Building overlays for n0002: [wwinit, generic]
   Building overlays for n0003: [wwinit, generic]
   Building overlays for n0004: [wwinit, generic]
   Building overlays for n0005: [wwinit, generic]
   Building overlays for n0006: [wwinit, generic]
   Building overlays for n0007: [wwinit, generic]
   Building overlays for n0008: [wwinit, generic]
   Building overlays for n0009: [wwinit, generic]
   Building overlays for n0010: [wwinit, generic]

Other Overlay Actions
=====================

Warewulf includes a number of overlay action commands to interact with the overlays in a programmatic and controlled manner. All of the commands use very similar usage structure and work as the above examples do. A summary of all of the overlay actions are as follows:

* **build**: (Re)build an overlay
* **chmod**: Change file permissions within an overlay
* **chown**: Change file ownership within an overlay
* **create**: Initialize a new Overlay
* **delete**: Delete Warewulf Overlay or files
* **edit**: Edit/Create a file within a Warewulf Overlay
* **import**: Import a file into a Warewulf Overlay
* **list**: List Warewulf Overlays and files
* **mkdir**: Create a new directory within an Overlay
* **show**: Show (cat) a file within a Warewulf Overlay