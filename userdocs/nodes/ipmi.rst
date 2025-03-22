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

Additionally, a ``vlan`` ipmi tag can be used to set the IPMI VLAN ID.

.. code-block::

   wwctl profile set default \
     --ipmitagadd vlan=42

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

   {{ $cmd := "ipmitool" }}
   {{ if .Interface }}{{ $cmd = cat $cmd "-I" .Interface }}{{ end }}
   {{ if .EscapeChar }}{{ $cmd = cat $cmd "-e" .EscapeChar }}{{ end }}
   {{ if .Port }}{{ $cmd = cat $cmd "-p" .Port }}{{ end }}
   {{ if .Ipaddr }}{{ $cmd = cat $cmd "-H" .Ipaddr }}{{ end }}
   {{ if .UserName }}{{ $cmd = cat $cmd "-U" (printf "\"%s\"" .UserName) }}{{ end }}
   {{ if .Password }}{{ $cmd = cat $cmd "-P" (printf "\"%s\"" .Password) }}{{ end }}
   {{ if eq .Cmd "PowerOn" }}{{ $cmd = cat $cmd "chassis power on" }}
   {{ else if eq .Cmd "PowerOff" }}{{ $cmd = cat $cmd "chassis power off" }}
   {{ else if eq .Cmd "PowerCycle" }}{{ $cmd = cat $cmd "chassis power cycle" }}
   {{ else if eq .Cmd "PowerReset" }}{{ $cmd = cat $cmd "chassis power reset" }}
   {{ else if eq .Cmd "PowerSoft" }}{{ $cmd = cat $cmd "chassis power soft" }}
   {{ else if eq .Cmd "PowerStatus" }}{{ $cmd = cat $cmd "chassis power status" }}
   {{ else if eq .Cmd "SDRList" }}{{ $cmd = cat $cmd "sdr list" }}
   {{ else if eq .Cmd "SensorList" }}{{ $cmd = cat $cmd "sensor list" }}
   {{ else if eq .Cmd "Console" }}{{ $cmd = cat $cmd "sol activate" }}
   {{ end }}
   {{- $cmd -}}

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
