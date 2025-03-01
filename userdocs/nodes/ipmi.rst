.. _ipmi:

====
IPMI
====

Warewulf can use IPMI to control cluster node power state or to connect to a
serial console.

Configuration
=============

Typically, common settings for IPMI interfaces are set on a profile, leaving
only the IP address set per-node.

If ``--ipmiwrite`` is set to `true`, the ``wwinit`` overlay will write the
desired IPMI configuration to the node's BMC during boot.

.. code-block::

    wwctl profile set default \
      --ipminetmask=255.255.255.0 \
      --ipmiuser=admin \
      --ipmipass=passw0rd \
      --ipmiinterface=lanplus \
      --ipmiwrite

    wwctl node set n1 \
      --ipmiaddr=192.168.2.1

``wwctl node list`` has a specific overview for IPMI settings.

.. code-block:: console

 # wwctl node list --ipmi
 NODE  IPMI IPADDR   IPMI PORT  IPMI USERNAME  IPMI INTERFACE
 ----  -----------   ---------  -------------  --------------
 n1    192.168.1.11  --         hwadmin        lanplus
 n2    192.168.1.12  --         hwadmin        lanplus
 n3    192.168.1.13  --         hwadmin        lanplus
 n4    192.168.1.14  --         hwadmin        lanplus

Power
=====

The ``wwctl power`` command can query and set the current power state of cluster
nodes.

.. code-block:: console
    
    wwctl power status n1 # query the current power status
    wwctl power off n1 # power off a cluster node
    wwctl power on n1 # power on a cluster node
    wwctl power reset n1 # forcibly reboot a node
    wwctl power soft n1 # ask a node to shut down gracefully
    wwctl power cycle n1 # power a cluster node off, then back on

Node ranges are supported; e.g., ``n[1-10]``.

Console
=======

If your node is setup to use serial over lan (SOL), Warewulf can connect a
console to the node.

.. code-block:: console

   # wwctl node console n1

Customization
=============

Warewulf doesn't manage IPMI interfaces directly, but uses ``ipmitool``. This is
configured with a template which defines Warewulf's IPMI behavior.

.. code-block::

   {{/* used command to access the ipmi interface of the nodes */}}
   {{- $escapechar := "~" }}
   {{- $port := "623" }}
   {{- $interface := "lan" }}
   {{- $args := "" }}
   {{- if .EscapeChar }} $escapechar = .EscapeChar {{ end }}
   {{- if .Port }} {{ $port = .Port }} {{ end }}
   {{- if .Interface }} {{ $interface = .Interface }} {{ end }}
   {{- if eq .Cmd "PowerOn" }} {{ $args = "chassis power on" }} {{ end }}
   {{- if eq .Cmd "PowerOff" }} {{ $args = "chassis power off" }} {{ end }}
   {{- if eq .Cmd "PowerCycle" }} {{ $args = "chassis power cycle" }} {{ end }}
   {{- if eq .Cmd "PowerReset" }} {{ $args = "chassis power reset" }} {{ end }}
   {{- if eq .Cmd "PowerSoft" }} {{ $args = "chassis power soft" }} {{ end }}
   {{- if eq .Cmd "PowerStatus" }} {{ $args = "chassis power status" }} {{ end }}
   {{- if eq .Cmd "SDRList" }} {{ $args = "sdr list" }} {{ end }}
   {{- if eq .Cmd "SensorList" }} {{ $args = "sensor list" }} {{ end }}
   {{- if eq .Cmd "Console" }} {{ $args = "sol activate" }} {{ end }}
   {{- $cmd := printf "ipmitool -I %s -H %s -p %s -U %s -P %s -e %s %s" $interface .Ipaddr $port .UserName .Password  $escapechar $args }}
   {{ $cmd }}

A different template can be used to change the IPMI behavior using the
``--ipmitemplate`` field. Referenced templates must be located in
``warewulf.conf:Paths.Datadir`` (``/usr/lib/warewulf/bmc/``).

All IPMI specific fields are accessible in the template:

+---------------------+--------------------+
| Parameter           | Template variable  |
+=====================+====================+
| ``--ipmiaddr``      | ``.Ipaddr``        |
+---------------------+--------------------+
| ``--ipminetmask``   | ``.Netmask``       |
+---------------------+--------------------+
| ``--ipmiport``      | ``.Port``          |
+---------------------+--------------------+
| ``--ipmigateway``   | ``.Gateway``       |
+---------------------+--------------------+
| ``--ipmiuser``      | ``.UserName``      |
+---------------------+--------------------+
| ``--ipmipass``      | ``.Password``      |
+---------------------+--------------------+
| ``--ipmiinterface`` | ``.Interface``     |
+---------------------+--------------------+
| ``--ipmiwrite``     | ``.Write``         |
+---------------------+--------------------+
| ``--ipmiescapechar``| ``.EscapeChar``    |
+---------------------+--------------------+
| ``--ipmitemplate``  | ``.Template``      |
+---------------------+--------------------+

Additionally, the ``.Cmd`` variable includes the relevant ``wwctl power``
subcommand.

* ``PowerOn``
* ``PowerOff``
* ``PowerCycle``
* ``PowerReset``
* ``PowerSoft``
* ``PowerStatus``
* ``SDRList``
* ``SensorList``
* ``Console``
