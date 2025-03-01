========
Overlays
========

Warewulf supplements provisioned node images with an "overlay" system. Overlays
are collections of files and :ref:`templates` that are rendered
and built per-node and then applied over the image during the provisioning
process.

Overlays are the primary mechanism for adding functionality Warewulf. Much of
even core functionality in Warewulf is implemented as distribution overlays, and
this flexibility is also available for local, custom overlays. By combining
templates with tags, network tags, and resources, the node registry
(``nodes.conf``) can become an expressive metadata store for arbitrary cluster
node configuration.

You can list the available overlays with ``wwctl overlay list``, and the files
within the overlays with ``wwctl overlay list --all``.

.. code-block:: console

   # wwctl overlay list --all fstab
   OVERLAY NAME  FILES/DIRS    SITE
   ------------  ----------    ----
   fstab         etc/          false
   fstab         etc/fstab.ww  false

Structure
=========

An overlay is a directory that is applied to the root of a cluster node's
runtime file system. The overlay source directory should contain a single
``rootfs`` directory which represents the actual root directory for the overlay.

.. code-block:: none

  /usr/share/warewulf/overlays/issue
  └── rootfs
      └── etc
          └── issue.ww

Adding Overlays to Nodes
========================

A node or profile can configure an overlay in two different ways:

* An overlay can be configured to apply only during boot, along with the node
  image. These overlays are called **system overlays**.
* An overlay can be configured to also apply periodically while the system is
  running. These overlays are called **runtime overlays**.

.. code-block:: shell

   wwctl profile set default \
     --system-overlays="wwinit,,wwclient,fstab,hostname,ssh.host_keys,systemd.netname,NetworkManager" \
     --runime-overlays="hosts,ssh.authorized_keys"

Multiple overlays can be applied to a single node, and overlays from multiple
profiles are appended together when applied to a single node.

Building Overlays
=================

Overlays are built (e.g., with ``wwctl overly build``) into compressed overlay
images for distribution to cluster nodes. These images typically match these two
use cases: system and runtime. As such, each cluster node typically has two
overlay images.

.. code-block:: console

   # wwctl overlay build
   Building system overlay image for n1
   Created image for n1 system overlay: /var/lib/warewulf/provision/overlays/n1/__SYSTEM__.img
   Compressed image for n1 system overlay: /var/lib/warewulf/provision/overlays/n1/__SYSTEM__.img.gz
   Building runtime overlay image for n1
   Created image for n1 runtime overlay: /var/lib/warewulf/provision/overlays/n1/__RUNTIME__.img
   Compressed image for n1 runtime overlay: /var/lib/warewulf/provision/overlays/n1/__RUNTIME__.img.gz

Overlay images for multiple node are built in parallel. By default, each CPU in
the Warewulf server will build overlays independently. The number of workers can
be specified with the ``--workers`` option.

Warewulf will attempt to build/update overlays as needed (configurable in the
``warewulf.conf``); but not all cases are detected, and manual overlay builds
are often necessary.

Creating and Modifying Overlays
===============================

You can add a new overlay to Warewulf with ``wwctl overlay create``.

.. code-block:: shell

   wwctl overlay create issue

A new overlay is just an empty directory. For it to be useful it needs to
contain some files.

For example, ``wwctl overlay import`` imports files from the Warewulf server
into the overlay.

.. code-block:: shell

   wwctl overlay import --parents issue /etc/issue

This imports ``/etc/issue`` from the Warewulf server into the new ``issue``
overlay.

.. note::

   The ``issue`` overlay already existed as a distribution overlay. Creating one
   shadows the distribution overlay with a new site overlay, allowing for local
   modification.

   Any modification to a distribution overlay first transparently creates a new
   site overlay and applies any changes there: distribution overlays should
   always remain unmodified.

You can also edit a new or existing overlay file in an interactive editor.

.. code-block:: shell

   wwctl overlay edit issue /etc/issue

Use ``wwctl overlay show`` to inspect the content of an overlay file.

.. code-block:: shell

   wwctl overlay show issue /etc/issue

Overlay files that end with ``.ww`` are templates. You can use ``wwctl overlay
show --render=<node>`` to show how a given template file would be rendered for
distribution to a given cluster node.

.. code-block:: shell

   wwctl overlay delete issue /etc/issue
   wwctl overlay import issue /etc/issue /etc/issue.ww
   wwctl overlay show issue /etc/issue.ww --render=n1

More information about templates is available in :ref:`its own section
<templates>`.

The content of the file for the given overlay is displayed with this command.
With the ``--render`` option a template is rendered as it will be rendered for
the given node. The node name is a mandatory argument to the ``--render`` flag.
Additional information for the file can be suppressed with the ``--quiet``
option.

.. note::

   It is not possible to delete files with an overlay.

Permissions
-----------

Overlay files are distributed to cluster nodes with the same user, group, and
mode that they have on the Warewulf server. Use ``wwctl overlay chown`` and
``wwctl overlay chmod`` to adjust them as necessary.

.. code-block:: shell

   wwctl overlay chown issue /etc/issue.ww root root
   wwctl overlay chmod issue /etc/issue.ww 0644

Distribution Overlays
=====================

Warewulf distinguishes between **distribution** overlays, which are included
with Warewulf, and **site** overlays, which are created or added locally. A site
overlay always takes precedence over a distribution overlay with the same name.
Any modification of a distribution overlay with ``wwctl`` actually makes changes
to an automatically-generated **site** overlay cloned from the distribution
overlay.

Site overlays are often stored at ``/var/lib/warewulf/overlays/``. Distribution
overlays are often stored at ``/usr/share/warewulf/overlays/``. But these paths
are dependent on compilation, distribution, packaging, and configuration
settings.

wwinit
------

The **wwinit** overlay performs initial configuration of the Warewulf node. Its
`wwinit` script runs before ``systemd`` or other init is called and contains all
configurations which are needed to boot.

In particular:

- Configure the loopback interface
- Configure the BMC based on the node's configuration
- Update PAM configuration to allow missing shadow entries
- Relabel the file system for SELinux

Other overlays may place additional scripts in ``/warewulf/init.d/`` to affect
node configuration in this pre-boot environment.

wwclient
--------

All configured overlays are provisioned initially along with the node image
itself; but **wwclient** periodically fetches and applies the runtime overlay to
allow configuration of some settings without a reboot.

Network interfaces
------------------

Warewulf ships with support for many different network interface configuration
systems. All of these are applied by default; but the list may be trimmed to
the desired system.

- ifcfg
- NetworkManager
- debian.interfaces
- wicked

Warewulf also configures both systemd and udev with the intended names of
configured network interfaces, typically based on a known MAC address.

- systemd.netname
- udev.netname

Several of the network configuration overlays support netdev tags to further
customize the interface:

- **``DNS[0-9]*``:** one or more DNS servers
- **``DNSSEARCH``:** domain search path
- **``MASTER``:** the master for a bond interface

NetworkManager
^^^^^^^^^^^^^^

- **``parent_device``:** the parent device of a vlan interface
- **``vlan_id``:** the vlan id for a vlan interface
- **``downdelay``, ``updelay``, ``miimon``, ``mode``, ``xmit_hash_policy``:**
  bond device settings

Basics
------

The **hostname** overlay sets the hostname based on the configured Warewulf
node name.

The **hosts** overlay configures ``/etc/hosts`` to include all Warewulf nodes.

The **issue** overlay configures a standard Warewulf status message for display
during login.

The **resolv** overlay configures ``/etc/resolv.conf`` based on the value of
"DNS" nettags. (In most situations this should be unnecessary, as the network
interface configuration should handle this dynamically.)

fstab
-----

The **fstab** overlay configures ``/etc/fstab`` based on the data provided in the "fstab"
resource. It also creates entries for file systems defined by Ignition.

.. code-block:: yaml

   nodeprofiles:
     default:
       resources:
         fstab:
           - spec: warewulf:/home
             file: /home
             vfstype: nfs
           - spec: warewulf:/opt
             file: /opt
             vfstype: nfs

ssh
---

Two SSH overlays configure host keys (one set for all node in the cluster) and
``authorized_keys`` for the root account.

- ssh.authorized_keys
- ssh.host_keys

syncuser
--------

The **syncuser** overlay updates ``/etc/passwd`` and ``/etc/group`` to include
all users on both the Warewulf server and from the image.

To function properly, ``wwctl image syncuser`` (or the ``--syncuser`` option
during ``import``, ``exec``, ``shell``, or ``build``) must have also been run on
the image to synchronize its user and group IDs with those of the server.

If a ``PasswordlessRoot`` tag is set to "true", the overlay will also insert a
"passwordless" root entry. This can be particularly useful for accessing a
cluster node when its network interface is not properly configured.

ignition
--------

The **ignition** overlay defines partitions and file systems on local disks.

debug
-----

The **debug** overlay is not intended to be used in configuration, but is
provided as an example. In particular, the provided `tstruct.md.ww` demonstrates
the use of most available template metadata.

.. code-block:: shell
  
   wwctl overlay show --render=<nodename> debug tstruct.md.ww

host
----

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
