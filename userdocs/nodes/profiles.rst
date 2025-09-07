=============
Node Profiles
=============

Node profiles provide a way to scalably group node configurations together.
Instead of redundant configurations for each node, you can set common fields in
a profile and then apply one or more profiles to each node.

Profiles may, themselves, reference other profiles, supporting complex mixtures
of profile configuration and negation.

The Default Profile
===================

A default Warewulf installation will come with a single "default" profile
pre-defined in ``nodes.conf``.


.. code-block:: console

   # wwctl profile list
   PROFILE NAME  COMMENT/DESCRIPTION
   ------------  -------------------
   default       This profile is automatically included for each node

If the default profile exists, each new node automatically includes it when it
is added.

You can view the fields of a profile with ``wwctl profile --all``.

.. code-block:: console

   # wwctl profile list default --all
   PROFILE  FIELD             VALUE
   -------  -----             -----
   default  Profiles          --
   default  Comment           This profile is automatically included for each node
   default  ClusterName       --
   default  ImageName         --
   default  Ipxe              default
   default  RuntimeOverlay    hosts,ssh.authorized_keys
   default  SystemOverlay     wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,ifupdown,NetworkManager,wicked,ignition
   default  Kernel.Version    --
   default  Kernel.Args       quiet,crashkernel=no
   default  Init              /sbin/init
   default  Root              initramfs
   default  PrimaryNetDev     --
   default  Resources[fstab]  [{"file":"/home","mntops":"defaults,nofail","spec":"warewulf:/home","vfstype":"nfs"},{"file":"/opt","mntops":"defaults,noauto,nofail,ro","spec":"warewulf:/opt","vfstype":"nfs"}]

``wwctl node list --all`` indicates which profile defines each field.

.. code-block:: console

   # wwctl node list n1 --all
   NODE  FIELD             PROFILE  VALUE
   ----  -----             -------  -----
   n1    Profiles          --       default
   n1    Comment           default  This profile is automatically included for each node
   n1    Ipxe              default  default
   n1    RuntimeOverlay    default  hosts,ssh.authorized_keys
   n1    SystemOverlay     default  wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,ifupdown,NetworkManager,wicked,ignition
   n1    Kernel.Args       default  quiet,crashkernel=no
   n1    Init              default  /sbin/init
   n1    Root              default  initramfs
   n1    Resources[fstab]  default  [{"file":"/home","mntops":"defaults,nofail","spec":"warewulf:/home","vfstype":"nfs"},{"file":"/opt","mntops":"defaults,noauto,nofail,ro","spec":"warewulf:/opt","vfstype":"nfs"}]

Setting Profile Fields
======================

(Almost) any node fields can be set on a profile, but some fields don't really
make sense anywhere but a node (e.g., ``--hwaddr`` and ``--ipaddr``).

.. code-block:: shell

   wwctl profile set default \
     --image=rockylinux-9 \
     --netmask=255.255.255.0

Multiple Profiles
=================

It's possible to create multiple profiles, and even to apply multiple profiles
to each node.

.. code-block:: shell

   wwctl profile add net
   wwctl profile set net --netmask=255.255.255.0

   wwctl profile add image
   wwctl profile set image --image=rockylinux-9

   wwctl node set n1 --profile="default,net,image"

.. note::

   If two profiles set the same field, the right-most profile in the node's list
   takes precedence. Field values set directly on nodes take precedence over
   profile field values.

.. code-block:: console

   # wwctl node list n1 --all
   NODE  FIELD                     PROFILE  VALUE
   ----  -----                     -------  -----
   n1    Profiles                  --       default,net,image
   n1    Comment                   default  This profile is automatically included for each node
   n1    ImageName                 image    rockylinux-9
   n1    Ipxe                      default  default
   n1    RuntimeOverlay            default  hosts,ssh.authorized_keys
   n1    SystemOverlay             default  wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,ifupdown,NetworkManager,wicked,ignition
   n1    Kernel.Args               default  quiet,crashkernel=no
   n1    Init                      default  /sbin/init
   n1    Root                      default  initramfs
   n1    NetDevs[default].Netmask  net      255.255.255.0
   n1    Resources[fstab]          default  [{"file":"/home","mntops":"defaults,nofail","spec":"warewulf:/home","vfstype":"nfs"},{"file":"/opt","mntops":"defaults,noauto,nofail,ro","spec":"warewulf:/opt","vfstype":"nfs"}]

Using multiple profiles makes it easy to work with multiple, heterogeneous
groups of cluster nodes and to test new configurations on smaller subsets of
nodes. For example, you can use this method to run a different kernel on only a
subset or group of cluster nodes without changing any other node attributes.

Negating Profiles
=================

Profiles may be negated by later profiles. For example, a profile list
``p2,~p1`` adds the profile ``p2`` to a node and removes a previously-applied
``p1`` profile from a node.

Using Profiles Effectively
==========================

There are a lot of ways to use profiles to facilitate complex cluster
configurations; but they are not required. It is completely possible to not use
profiles at all, and to simply set all fields directly on cluster nodes.

If you do use profiles, some fields lend themselves most naturally to being set
on profiles. Network subnet masks (``--netmask``) and gateways (``--gateway``)
are common profile fields, as is ``--image``. Most :ref:`IPMI <ipmi>` fields
make sense on a profile, and it is also common to configure tags and resources
on a profile for easy application to multiple nodes.

Node-specific information, like HW/MAC addresses (``--hwaddr``) and IP addresses
(``--ipaddr``, ``--ipmiaddr``) should always be put in a node configuration
rather than a profile configuration.
