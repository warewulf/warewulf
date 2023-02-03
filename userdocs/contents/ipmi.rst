====
IPMI
====

It is possible to control the power or connect a console to your nodes being managed by Warewulf by connecting to the BMC through the use of IPMI. We will discuss how to set this up below.

IPMI Settings
=============

The common settings for the IPMI interfaces on all nodes can be set on a Profile level.  The only field that would need to be assigned to each individual node would be the IP address.

If an individual node has different settings, you can set the IPMI settings for that specific node, overriding the default settings.

Here is a table outlining the fields on a Profile and Node level, along with the parameters that can be used when running ``wwctl profile set`` or ``wwctl node set``.

============= =============== ======== ======= ================== =============
Field         Parameter       Profile  Node    Valid Values       Default Value
============= =============== ======== ======= ================== =============
IpmiIpaddr    --ipmi                   X
IpmiNetmask   --ipminetmask    X       X
IpmiPort      --ipmiport       X       X                          623
IpmiGateway   --ipmigateway    X       X
IpmiUserName  --ipmiuser       X       X
IpmiPassword  --ipmipass       X       X
IpmiInterface --ipmiinterface  X       X       'lan' or 'lanplus' lan
============= =============== ======== ======= ================== =============

Reviewing Settings
==================

There are multiple ways to review the IPMI settings. They can be reviewed from a Profile level, all the way down to a specific Node.

Profile View
------------

.. code-block:: bash

   $ sudo wwctl profile list -a

   ################################################################################
   PROFILE NAME         FIELD              VALUE
   default              Id                 default
   default              Comment            This profile is automatically included for each node
   default              Cluster            --
   default              Container          rocky
   default              Kernel             4.18.0-348.2.1.el8_5.x86_64
   default              KernelArgs         --
   default              Init               --
   default              Root               --
   default              RuntimeOverlay     --
   default              SystemOverlay      --
   default              Ipxe               --
   default              IpmiNetmask        255.255.255.0
   default              IpmiPort           --
   default              IpmiGateway        192.168.99.1
   default              IpmiUserName       admin
   default              IpmiInterface      lanplus
   default              eth0:IPADDR        --
   default              eth0:NETMASK       255.255.240.0
   default              eth0:GATEWAY       10.1.96.6
   default              eth0:HWADDR        --
   default              eth0:TYPE          --
   default              eth0:DEFAULT       false

Node View
---------

.. code-block:: bash

   $ sudo wwctl node list node0001      -a

   ################################################################################
   NODE                 FIELD              PROFILE      VALUE
   node0001             Id                 --           node0001
   node0001             Comment            default      This profile is automatically included for each node
   node0001             Cluster            --           --
   node0001             Profiles           --           default
   node0001             Discoverable       --           false
   node0001             Container          default      rocky
   node0001             Kernel             default      4.18.0-348.2.1.el8_5.x86_64
   node0001             KernelArgs         --           (quiet crashkernel=no vga=791 rootfstype=rootfs)
   node0001             RuntimeOverlay     --           (default)
   node0001             SystemOverlay      --           (default)
   node0001             Ipxe               --           (default)
   node0001             Init               --           (/sbin/init)
   node0001             Root               --           (initramfs)
   node0001             IpmiIpaddr         --           192.168.99.10
   node0001             IpmiNetmask        --           255.255.255.0
   node0001             IpmiPort           --           --
   node0001             IpmiGateway        --           192.168.99.1
   node0001             IpmiUserName       default      admin
   node0001             IpmiInterface      default      lanplus
   node0001             eth0:HWADDR        --           52:54:00:1a:08:60
   node0001             eth0:IPADDR        --           192.168.100.152
   node0001             eth0:NETMASK       default      255.255.255.0
   node0001             eth0:GATEWAY       default      192.168.100.1
   node0001             eth0:TYPE          --           --
   node0001             eth0:DEFAULT       --           false

Review Only IPMI Settings
-------------------------

The above views show you everything that is set on a Profile or Node level. That is a lot of detail. If you want to view key IPMI connecton details for a node or all nodes, you can do the following.

.. code-block:: bash

   $ sudo wwctl node list -i

   NODE NAME              IPMI IPADDR      IPMI PORT  IPMI USERNAME        IPMI PASSWORD        IPMI INTERFACE
   ============================================================================================================
   node0001               192.168.99.10    --         admin                supersecret          lanplus
   node0002               192.168.99.11    --         admin                supersecret          lanplus
   node0003               192.168.99.12    --         admin                supersecret          lanplus

Power Commands
==============

The ``power`` command can be used to manage the current power state of your nodes through IPMI.

``wwctl power [command]`` where ``[command]`` is one of:

cycle
    Turns the power off and then on

off
    Turns the power off

on
    Turns the power on

reset
    Issues a reset

soft
    Shutdown gracefully

status
    Shows current power status

Console
=======

If your node is setup to use serial over lan (SOL), Warewulf can connect a console to the node.

``sudo wwctl node console node0001``