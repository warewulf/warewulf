=============
Node Profiles
=============

Profiles provide a way to scalably group node configurations
together. Instead of redundant configurations for each node, you can
put that into a profile and the nodes will inherit these
configurations. This is very handy if you have groups of node specific
customizations. For example, a few hundred nodes that are running a
particular container or kernel, and another group of nodes that are
running a different kernel or container.

Any node configuration attributes can be applied to a profile, but
there are always going to be some node configurations which must be
specific to a node, like a network HW/MAC address or an IP address.

An Introduction To Profiles
===========================

Every new node is automatically added to a profile called
``default``. You can view the configuration attributes of this profile
by using the ``wwctl profile list`` command. Like the ``wwctl node
list`` command, this will provide a summary, but you can see **all**
configuration attributes by using the ``--all`` or ``-a`` flag as
follows:

.. code-block:: console

   # wwctl profile list
   PROFILE NAME         COMMENT/DESCRIPTION
   ================================================================================
   default              This profile is automatically included for each node

And with the ``-a`` flag:

.. code-block:: console

   # wwctl profile list -a
   ################################################################################
   PROFILE NAME         FIELD              VALUE
   default              Id                 default
   default              Comment            This profile is automatically included for each node
   default              Cluster            --
   default              Container          --
   default              Kernel             --
   default              KernelArgs         --
   default              Init               --
   default              Root               --
   default              RuntimeOverlay     --
   default              SystemOverlay      --
   default              Ipxe               --
   default              IpmiIpaddr         --
   default              IpmiNetmask        --
   default              IpmiPort           --
   default              IpmiGateway        --
   default              IpmiUserName       --
   default              IpmiInterface      --

As you can see here, there is only one attribute set by default in
this profile, and that is the "Comment" field. That Comment is
inherited by any nodes that are configured to use this profile. So if
we look at the node we configured in the last section, we can see that
configuration attribute there:

.. code-block:: console

   # wwctl node list -a | head -n 5
   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   n0000                Id                 --           n0000
   n0000                Comment            default      This profile is automatically included for each node
   n0000                Cluster            --           --

Here you can see that the "Comment" attribute was inherited by this
node, and it also provides you with the information of which profile
this attribute was inherited from. This is very useful information as
nodes can be part of multiple profiles with inheritance being
cascading.

Multiple Profiles
=================

For demonstration purposes, let's create another profile and
demonstrate how to use this second profile.

.. code-block:: console

   # wwctl profile add test_profile
   # wwctl profile list
   PROFILE NAME         COMMENT/DESCRIPTION
   ================================================================================
   default              This profile is automatically included for each node
   test_profile         --

Now that we've created a new profile, let's create a configuration
attribute in this profile:

.. code-block:: console

   # wwctl profile set --cluster cluster01 test_profile
   ? Are you sure you want to modify 1 profile(s)? [y/N] y

   # wwctl profile list -a test_profile | grep Cluster
   test_profile         Cluster            cluster01

Lastly we just need to configure this profile to our node(s):

.. code-block:: console

   # wwctl node set --addprofile test_profile n0000
   Are you sure you want to modify 1 nodes(s): y

And you can now verify that the node has both profile configurations:

.. code-block:: console

   # wwctl node list -a | head -n 6
   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   n0000                Id                 --           n0000
   n0000                Comment            default      This profile is automatically included for each node
   n0000                Cluster            test_profile cluster01
   n0000                Profiles           --           default,test_profile

Cascading Profiles
==================

In the previous example, we set a single node to have two profile
configurations. We can also overwrite configurations as follows:

.. code-block:: console

   # wwctl profile set --comment "test comment" test_profile
   Are you sure you want to modify 1 profile(s): y

   # wwctl node list -a | head -n 6
   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   n0000                Id                 --           n0000
   n0000                Comment            test_profile test comment
   n0000                Cluster            test_profile cluster01
   n0000                Profiles           --           default,test_profile

And if we delete the superseded profile attribute from
``test_profile`` we can now see the previous configuration:

.. code-block:: console

   # wwctl profile set --comment UNDEF test_profile
   Are you sure you want to modify 1 profile(s): y

   # wwctl node list -a | head -n 6
   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   n0000                Id                 --           n0000
   n0000                Comment            default      This profile is automatically included for each node
   n0000                Cluster            test_profile cluster01
   n0000                Profiles           --           default,test_profile

This is a very useful feature for dealing with many groups of cluster
nodes and/or testing new configurations on smaller subsets of cluster
nodes. For example, you can use this method to run a different kernel
on only a subset or group of cluster nodes without changing any other
node attributes.

Overriding Profiles
===================

All profile configurations can be overwritten by a node configuration
as can be seen here:

.. code-block:: console

   # wwctl node set --comment "This value takes precedent" n0000
   Are you sure you want to modify 1 nodes(s): y

   # wwctl node list -a | head -n 6
   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   n0000                Id                 --           n0000
   n0000                Comment            SUPERSEDED   This value takes precedent
   n0000                Cluster            test_profile cluster01
   n0000                Profiles           --           default,test_profile

How To Use Profiles Effectively
===============================

There are a lot of ways to use profiles to facilitate the management
of large cluster node attributes, but there is nothing inherent in the
design of Warewulf that requires use of them for anything. It is
completely reasonable to not use profiles at all to help with node
configuration attributes.

But if you do wish to use profiles, the best way to use them is to
manage "fixed" configurations of groups of cluster nodes. For example,
if you have multiple sub-clusters in your cluster, it might be
advantageous to have a ``cluster_name`` profile which includes things
like network configurations, and/or a specific kernel, container, boot
arguments, etc.

Node specific information, like HW/MAC addresses and IP addresses
should always be put in a node configuration rather than a profile
configuration.
