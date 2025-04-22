=============
Cluster Nodes
=============

Warewulf cluster node configuration is persisted in ``nodes.conf`` (also known
as the "node registry" or "node database"). Editing this file directly is
supported; but it is often better to manage it using the ``wwctl`` command.

.. note::

   The ``nodes.conf`` file is YAML document that can be edited directly or
   managed with configuation management; but its internal structure is
   technically undocumented and subject to change between versions. After
   Warewulf v4.6.0, the ``wwctl upgrade nodes`` command can be used to update a
   ``nodes.conf`` from a previous Warewulf v4 version.

.. warning::
   
   When ``nodes.conf`` is edited directly, ``warewulfd`` must be restarted to reflect the changes.

   .. code-block:: shell

      systemctl restart warewulfd.service


Adding a Cluster Node
=====================

Adding a cluster node is as simple as running ``wwctl node add``.

.. code-block:: console

   # wwctl node add n1 --ipaddr=10.0.2.1
   Added node: n1

Several nodes can be added with a node range. In this case, the provided IP
address is automatically incremented.

.. code-block:: console

   # wwctl node add n[2-4] --ipaddr=10.0.2.2
   Added node: n2
   Added node: n3
   Added node: n4

   # wwctl node list --net n[1-4]
   NODE  NETWORK  HWADDR  IPADDR    GATEWAY  DEVICE
   ----  -------  ------  ------    -------  ------
   n1    default  --      10.0.2.1  <nil>    --
   n2    default  --      10.0.2.2  <nil>    --
   n3    default  --      10.0.2.3  <nil>    --
   n4    default  --      10.0.2.4  <nil>    --


Listing Nodes
=============

Once you have configured one or more nodes, you can list them and their
attributes with ``wwctl node list``.

.. code-block:: console

   # wwctl node list n[1-5]
   NODE NAME  PROFILES  NETWORK
   ---------  --------  -------
   n1         default   --
   n2         default   --
   n3         default   --
   n4         default   --
   n5         default   --

You can also see the node's full attribute list by specifying ``--all``.

.. code-block:: console

   # wwctl node list --all n1
   NODE  FIELD             PROFILE  VALUE
   ----  -----             -------  -----
   n1    Profiles          --       default
   n1    Comment           default  This profile is automatically included for each node
   n1    Ipxe              default  default
   n1    RuntimeOverlay    default  hosts,ssh.authorized_keys
   n1    SystemOverlay     default  wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,NetworkManager,debian.interfaces,wicked,ignition
   n1    Kernel.Args       default  quiet,crashkernel=no
   n1    Init              default  /sbin/init
   n1    Root              default  initramfs
   n1    Resources[fstab]  default  [{"file":"/home","mntops":"defaults,nofail","spec":"warewulf:/home","vfstype":"nfs"},{"file":"/opt","mntops":"defaults,noauto,nofail,ro","spec":"warewulf:/opt","vfstype":"nfs"}]


Setting Node Fields
===================

Node fields are set using the ``wwctl node set`` command. A list of all
available fields is available with ``wwctl node set --help``.

You can also edit nodes as YAML data in an interactive editor using ``wwctl node
edit``.

List values
-----------

Some node fields, such as overlays and kernel aruments, accept a list of values.
These may be specified as a comma-separated list or as multiple arguments.

To include an explicit comma in the value, enclose the value in inner-quotes.

.. code-block:: shell

   wwctl node set n1 \
     --kernelargs 'quiet,crashkernel=no,nosplash' \
     --kernelargs='"console=ttyS0,115200"'

Un-setting Node Fields
----------------------

To un-set a field value, set the value to ``UNDEF``.

.. code-block:: shell

   wwctl node set n1 \
     --image=UNDEF

Configuring an Image
====================

One of the main things to configure for a cluster node is the image that it
should provision.

.. code-block:: shell

   wwctl node set n1 \
     --image=rockylinux-9

Images are covered in more detail :ref:`in their own section. <images>`

Configuring the Network
=======================

By default, network configurations are applied to a "default" network interface.

.. code-block:: shell

  wwctl node set n1 \
    --netdev=eno1 \
    --hwaddr=00:00:00:00:00:01 \
    --ipaddr=10.0.2.1 \
    --netmask=255.255.255.0

Network interface configuration is covered in more detail :ref:`in its own
section. <node-network>`

Node Discovery
==============

The MAC / hardware address (``--hwaddr``) of a cluster node can be automatically
discovered by marking the node ``--discoverable``. If a node attempts to
provision against Warewulf using an interface that is unknown to Warewulf, its
hardware address becomes associated with the first discoverable node. (Multiple
discoverable nodes are sorted lexically, first by cluster, then by ID.)

Once a node has been discovered its "discoverable" field is automatically
cleared.

Tags
====

Cluster nodes support multiple key-value pair tags. Tags may be applied to the
node directly, to network interfaces, and even to IPMI interfaces.

.. code-block:: shell

   wwctl node set n1 --tagadd="localtime=UTC"
   wwctl node set n1 --nettagadd="DNS1=1.1.1.1"

Resources
=========

Cluster nodes support generic "resources" that may hold arbitrarily complex YAML
data. This data, along with tags, may be used by both distribution and site
overlays.

.. code-block:: yaml

   nodeprofiles:
     default:
       resources:
         fstab:
           - spec: warewulf:/home
             file: /home
             vfstype: nfs
             mntops: defaults
             freq: 0
             passno: 0
           - spec: warewulf:/opt
             file: /opt
             vfstype: nfs
             mntops: defaults,ro
             freq: 0
             passno: 0

Resources can only be managed with ``wwctl node edit``.

Importing Nodes From a File
===========================

You can import nodes into Warewulf by using the ``wwctl node import`` command. The
file used must be in YAML or CSV format. 

.. warning::
   Importing a node configuration will fully overwrite the existing settings, 
   including any customizations not present in the import file. If the node 
   already exists and you wish to update it, ensure that the import file 
   includes all the options you want to retain.

CSV Import
----------

.. note::
   As of Warewulf v4.6.1, the csv import functionality is broken and an 
   `issue <https://github.com/warewulf/warewulf/issues/1862>`_
   has been created to track this.

The CSV file must have a header where the first field must always be the nodename, 
and the rest of the fields are the same as the long commandline options. Network 
device must have the form ``net.$NETNAME.$NETOPTION``. (e.g., ``net.default.ipaddr``). 
Tags are currently not supported and must be added separately after the import. 

As an example, the following CSV file:

.. code-block:: csv

   nodename,net.default.hwaddr,net.default.ipaddr,net.default.netmask,net.default.gateway,discoverable,image
   n1,00:00:00:00:00:01,10.0.2.1,255.255.255.0,10.0.2.254,false,rockylinux-9

This can be imported with the following command:

.. code-block:: shell

   wwctl node import --csv /path/to/nodes.csv

YAML Import
-----------

The YAML file must be a mapping of node names to their attributes, where each node is represented as a dictionary of attributes. 
To simplify the creation of the YAML file, you can use the wwctl node export command to export the current node configuration to a YAML file. 
This exported file can serve as a template for creating new nodes.

A minimal example of a YAML file looks like this:

.. code-block:: yaml

   n1:
      profiles:
         - default
      image name: rockylinux-9
      ipxe template: default
      kernel:
         args:
            - quiet
            - crashkernel=no
            - nosplash
            - console=ttyS0,115200
      network devices:
         default:
            type: ethernet
            device: eno1
            hwaddr: "00:00:00:00:00:01"
            ipaddr: 10.0.2.1
            netmask: 255.255.255.0
            gateway: 172.16.131.1
            tags:
            DNS1: 1.1.1.1
      primary network: default

This can be imported with the following command:

.. code-block:: shell

   wwctl node import /path/to/nodes.yaml
