====
IPMI
====

It is possible to control the power or connect a console to your nodes being managed by Warewulf by connecting to the BMC through the use of IPMI. We will discuss how to set this up below.

IPMI Settings
=============

The common settings for the IPMI interfaces on all nodes can be set on a Profile level.  The only field that would need to be assigned to each individual node would be the IP address.

If an individual node has different settings, you can set the IPMI settings for that specific node, overriding the default settings.

Here is a table outlining the fields on a Profile and Node which is the same as the parameter that can be used when running `wwctl profile set` or `wwctl node set`.

+----------------+---------+------+--------------------+---------------+
| Parameter      | Profile | Node | Valid Values       | Default Value |
+================+=========+======+====================+===============+
| --ipmiaddr     | false   | true |                    |               |
+----------------+---------+------+--------------------+---------------+
| --ipminetmask  | true    | true |                    |               |
+----------------+---------+------+--------------------+---------------+
| --ipmiport     | true    | true |                    | 623           |
+----------------+---------+------+--------------------+---------------+
| --ipmigateway  | true    | true |                    |               |
+----------------+---------+------+--------------------+---------------+
| --ipmiuser     | true    | true |                    |               |
+----------------+---------+------+--------------------+---------------+
| --ipmipass     | true    | true |                    |               |
+----------------+---------+------+--------------------+---------------+
| --ipmiinterface| true    | true | 'lan' or 'lanplus' | lan           |
+----------------+---------+------+--------------------+---------------+
| --ipmiwrite    | true    | true | true or false      | false         |
+----------------+---------+------+--------------------+---------------+


Reviewing Settings
==================

There are multiple ways to review the IPMI settings. They can be reviewed from a Profile level, all the way down to a specific Node.

Profile View
------------

.. code-block:: bash

  $ sudo wwctl profile list -a
  PROFILE              FIELD              PROFILE      VALUE
  =====================================================================================
  default              Id                 --           default
  default              comment            --           This profile is automatically included for each node
  default              cluster            --           --
  default              container          --           sle-micro-5.3
  default              ipxe               --           --
  default              runtime            --           --
  default              wwinit             --           --
  default              root               --           --
  default              discoverable       --           --
  default              init               --           --
  default              asset              --           --
  default              profile            --           --
  default              default:type       --           --
  default              default:onboot     --           --
  default              default:netdev     --           --
  default              default:hwaddr     --           --
  default              default:ipaddr     --           --
  default              default:ipaddr6    --           --
  default              default:netmask    --           --
  default              default:gateway    --           --
  default              default:mtu        --           --
  default              default:primary    --           --


Node View
---------
.. code-block:: bash

  $ sudo wwctl node list -a n001
  NODE                 FIELD              PROFILE      VALUE
  =====================================================================================
  n001                 Id                 --           n001
  n001                 comment            default      This profile is automatically included for each node
  n001                 cluster            --           --
  n001                 container          default      sle-micro-5.3
  n001                 ipxe               --           (default)
  n001                 runtime            --           (generic)
  n001                 wwinit             --           (wwinit)
  n001                 root               --           (initramfs)
  n001                 discoverable       --           --
  n001                 init               --           (/sbin/init)
  n001                 asset              --           --
  n001                 kerneloverride     --           tw
  n001                 kernelargs         --           (quiet crashkernel=no vga=791 net.naming-scheme=v238)
  n001                 ipmiaddr           --           --
  n001                 ipminetmask        --           --
  n001                 ipmiport           --           --
  n001                 ipmigateway        --           --
  n001                 ipmiuser           --           --
  n001                 ipmipass           --           --
  n001                 ipmiinterface      --           --
  n001                 ipmiwrite          --           --
  n001                 profile            --           default
  n001                 default:type       --           (ethernet)
  n001                 default:onboot     --           --
  n001                 default:netdev     --           eth0
  n001                 default:hwaddr     --           11:22:33:44:55:66
  n001                 default:ipaddr     --           10.0.2.1
  n001                 default:ipaddr6    --           --
  n001                 default:netmask    --           255.255.252.0
  n001                 default:gateway    --           --
  n001                 default:mtu        --           --
  n001                 default:primary    --           true

Review Only IPMI Settings
-------------------------

The above views show you everything that is set on a Profile or Node level. That is a lot of detail. If you want to view key IPMI connecton details for a node or all nodes, you can do the following.

.. code-block:: bash

   $ sudo wwctl node list -i
 NODE NAME              IPMI IPADDR      IPMI PORT  IPMI USERNAME        IPMI INTERFACE
 ==================================================================================================
 n001                   192.168.1.11     --         hwadmin              --            
 n002                   192.168.1.12     --         hwadmin              --            
 n003                   192.168.1.13     --         hwadmin              --            
 n004                   192.168.1.14     --         hwadmin              --            


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

``sudo wwctl node console n001``
