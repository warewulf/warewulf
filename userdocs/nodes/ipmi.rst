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

If ``--ipmiwrite`` is set to ``true``, the ``wwinit`` overlay will write the
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

A ``bit-rate`` ipmi tag can be used to set the Serial over LAN bit rate (defaults to 38.4).
Typical options are 19.2, 38.4, and 115.2.

.. code-block::

   wwctl profile set default \
     --ipmitagadd bit-rate=115.2

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

Virtual BMC Templates
======================

Warewulf includes templates for managing virtual machines as simulated cluster
nodes for testing and development. These templates use IPMI tags to configure
virtual machine resources like CPU, memory, and disk size.

Available Virtual BMC Templates
--------------------------------

kind.tmpl (Docker-based)
~~~~~~~~~~~~~~~~~~~~~~~~

Uses Docker containers to simulate virtual nodes. Lightweight and doesn't require
full hardware virtualization.

.. code-block:: console

   wwctl node set vnode1 \
     --ipmiaddr=10.0.0.100 \
     --ipmitemplate=kind.tmpl \
     --ipmitagadd nodename=vnode1 \
     --ipmitagadd cpu=4 \
     --ipmitagadd memory=4g \
     --ipmitagadd disk=50g \
     --ipmitagadd image=kindest/node:latest

kind-libvirt.tmpl (KVM/Libvirt-based)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Uses libvirt and KVM for full hardware virtualization. Provides more realistic
virtual machines with better isolation.

.. code-block:: console

   wwctl node set vnode2 \
     --ipmiaddr=10.0.0.101 \
     --ipmitemplate=kind-libvirt.tmpl \
     --ipmitagadd nodename=vnode2 \
     --ipmitagadd cpu=4 \
     --ipmitagadd memory=4096 \
     --ipmitagadd disk=50 \
     --ipmitagadd disk_path=/var/lib/libvirt/images \
     --ipmitagadd os_variant=ubuntu22.04

kind-qemu.tmpl (QEMU-based)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Uses QEMU directly for virtual machine management. Provides fine-grained control
over VM configuration and doesn't require libvirt.

.. code-block:: console

   wwctl node set vnode3 \
     --ipmiaddr=10.0.0.102 \
     --ipmitemplate=kind-qemu.tmpl \
     --ipmitagadd nodename=vnode3 \
     --ipmitagadd cpu=2 \
     --ipmitagadd memory=2048 \
     --ipmitagadd disk=20 \
     --ipmitagadd disk_path=/var/lib/qemu/images \
     --ipmitagadd mac=52:54:00:12:34:56

Virtual BMC Configuration Tags
-------------------------------

The following IPMI tags can be used to configure virtual machine resources:

+-------------------+------------------------------------+-----------------------------+
| Tag               | Description                        | Default                     |
+===================+====================================+=============================+
| ``nodename``      | Custom VM name                     | IP address with dashes      |
+-------------------+------------------------------------+-----------------------------+
| ``cpu``           | Number of CPUs                     | 2                           |
+-------------------+------------------------------------+-----------------------------+
| ``memory``        | Memory size                        | 2g (kind), 2048 (others)    |
+-------------------+------------------------------------+-----------------------------+
| ``disk``          | Disk size                          | 20g (kind), 20 (others)     |
+-------------------+------------------------------------+-----------------------------+
| ``disk_path``     | Path for VM disk images            | /var/lib/libvirt/images or  |
|                   |                                    | /var/lib/qemu/images        |
+-------------------+------------------------------------+-----------------------------+
| ``image``         | Docker image (kind.tmpl only)      | kindest/node:latest         |
+-------------------+------------------------------------+-----------------------------+
| ``os_variant``    | OS variant (kind-libvirt.tmpl)     | generic                     |
+-------------------+------------------------------------+-----------------------------+
| ``mac``           | MAC address (kind-qemu.tmpl)       | Auto-generated              |
+-------------------+------------------------------------+-----------------------------+

Virtual Node Power Management
------------------------------

Virtual nodes can be managed using the same ``wwctl power`` commands as physical
nodes:

.. code-block:: console

   # Power on creates the VM if it doesn't exist
   wwctl power on vnode1

   # Power status checks if VM is running
   wwctl power status vnode1

   # SDR list shows VM configuration
   wwctl power sdr vnode1

   # Power off stops the VM but doesn't destroy it
   wwctl power off vnode1

**PowerOn Behavior**: When powering on a virtual node, the template will:

1. Check if the VM already exists
2. If it exists and is stopped, start it
3. If it doesn't exist, create it with the configured resources, then start it

**PowerOff Behavior**: Powering off stops the VM but preserves it for future use.
The VM and its disk are not destroyed.

Template Access
---------------

IPMI tags can be accessed in templates using the ``.Tags`` map:

.. code-block::

   {{ $cpu := "2" }}{{ if .Tags.cpu }}{{ $cpu = .Tags.cpu }}{{ end }}
   {{ $memory := "2048" }}{{ if .Tags.memory }}{{ $memory = .Tags.memory }}{{ end }}
   {{ $nodeName := .Ipaddr | toString | replace "." "-" }}
   {{ if .Tags.nodename }}{{ $nodeName = .Tags.nodename }}{{ end }}
