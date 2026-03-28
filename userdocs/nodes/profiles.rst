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

Profile Precedence and Combinations of Profiles
=============================================

In case of scalar fields, the right-most profile in the node's profile list
takes precedence. Field values set directly on nodes take precedence over
profile field values. However, things behave differently for field types, like
lists, where it makes sense to combine their values.

Let's assume we have in addition to the node profile two more profiles:
``default`` and ``profile_a``. The field ``root`` serves as an example
for a scalar field and ``runtime overlay`` for list typed fields.

We start with a configuration where the node uses only the default profile.

.. code-block:: console

   nodeprofiles:
      profile_a:
        runtime overlay:
          - runtime_overlay_from_profile_a
        root: root_from_profile_a
      default:
        runtime overlay:
          - runtime_overlay_from_default
        root: root_from_default
   nodes:
     n1:
       profiles:
         - default
       runtime overlay:
         - runtime_overlay_from_node_profile
       root: root_from_node_profile

Next we filter the output a little to focus on one field at time.

.. code-block:: console

   # wwctl node list n1 --all | grep -E '(NODE|----|Root)'
     NODE  FIELD             PROFILE               VALUE
     ----  -----             -------               -----
     n1    Root              SUPERSEDED            root_from_node_profile

As expected, the value originates from the node profile because the node
profile has highest precedence.

Things are different for the list typed field ``runtime overlay``.

.. code-block:: console

   #  wwctl node list n1 --all | grep -E '(NODE|----|Runtime)'
      NODE  FIELD             PROFILE     VALUE
      ----  -----             -------     -----
      n1    RuntimeOverlay    default,n1  runtime_overlay_from_default,runtime_overlay_from_node_profile

Here the value from the node profile is appended to the value of the
default profile.

Next we remove the value for ``root`` from the node profile and add
``profile_a`` to the list of profiles for the node.

.. code-block:: shell

   nodes:
     n1:
       profiles:
         - default
         - profile_a
       runtime overlay:
         - runtime_overlay_from_node_profile

First we check the ``root`` field and see that it set from ``profile_a``
as ``profile_a`` is now the profile with highest precedence.

.. code-block:: console

   # wwctl node list n1 --all | grep -E '(NODE|----|Root)'
     NODE  FIELD             PROFILE               VALUE
     ----  -----             -------               -----
     n1    Root              profile_a             root_from_profile_a

The value of ``runtime overlay`` is now a combination of the values of
all three profiles.

.. code-block:: console

   # wwctl node list n1 --all | grep -E '(NODE|----|Runtime)'
   NODE  FIELD             PROFILE               VALUE
   ----  -----             -------               -----
   n1    RuntimeOverlay    default,profile_a,n1  runtime_overlay_from_default,runtime_overlay_from_profile_a,runtime_overlay_from_node_profile

We get the same result when making use of nested profiles, that is
when we add ``profile_a`` to the list of profiles within the default profile.

.. code-block:: shell

   nodeprofiles:
      default:
        profiles:
          - profile_a
        ...
      ...
   nodes:
     n1:
       profiles:
         - default
       ...

Negating Profiles
=================

Profiles may be negated by later profiles. For example, a profile list
``p2,~p1`` adds the profile ``p2`` to a node and removes a previously-applied
``p1`` profile from a node.

.. code-block:: console

   nodeprofiles:
     p1:
       runtime overlay:
         - runtime_overlay_from_p1
     p2:
       profiles:
         - p1
       runtime overlay:
         - runtime_overlay_from_p2
   nodes:
     n1:
       profiles:
         - p2
         - ~p1

The value of ``runtime overlay`` is then just ``runtime_overlay_from_p2``.

.. code-block:: shell

   # wwctl node list n2 --all
   NODE  FIELD           PROFILE  VALUE
   ----  -----           -------  -----
   n2    Profiles        --       p2,~p1
   n2    RuntimeOverlay  p2       runtime_overlay_from_p2

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
