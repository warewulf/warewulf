==========
Networking
==========

Multiple networks
=================

It is possible to configure several networks not just for the nodes but also for
the management of ``dhcpd`` and ``tftp``. There are two ways to achieve this:

* Add the networks to the templates of ``dhcpd`` and/or the ``dnsmasq`` template
  directly.
* Add the networks to a dummy node and change the templates of ``dhcp`` and
  ``dnsmasq`` accordingly.

The first method is relatively trivial. The second method is described below.

As the first step, add the dummy node.

.. code-block:: shell

   wwctl node add deliverynet

Add the delivery networks to this node.

.. code-block:: shell

  wwctl node set \
    --ipaddr 10.0.20.250 \
    --netmask 255.255.255.0 \
    --netname deliver1 \
    --nettagadd network=10.0.20.0,dynstart=10.10.20.10,dynend=10.10.20.50 \
    deliverynet

  wwctl node set \
    --ipaddr 10.0.30.250 \
    --netmask 255.255.255.0 \
    --netname deliver2 \
    --nettagadd network=10.0.30.0,dynstart=10.10.30.10,dynend=10.10.30.50 \
    deliverynet

The IP address is used as the network address of the host in the delivery network
and an additional tag is used for the definition of the network itself and the
dynamic dhcp range. You can check the result with ``wwctl node list``.

.. code-block:: console

  # wwctl node list -a deliverynet
  NODE         FIELD                             PROFILE  VALUE
  deliverynet  Id                                --       deliverynet
  deliverynet  Comment                           default  This profile is automatically included for each node
  deliverynet  ImageName                         default  leap15.5
  deliverynet  Ipxe                              --       (default)
  deliverynet  RuntimeOverlay                    --       (hosts,ssh.authorized_keys)
  deliverynet  SystemOverlay                     --       (wwinit,wwclient,hostname,ssh.host_keys,systemd.netname,NetworkManager)
  deliverynet  Root                              --       (initramfs)
  deliverynet  Init                              --       (/sbin/init)
  deliverynet  Kernel.Args                       --       (quiet crashkernel=no net.ifnames=1)
  deliverynet  Profiles                          --       default
  deliverynet  PrimaryNetDev                     --       (deliver1)
  deliverynet  NetDevs[deliver2].Type            --       (ethernet)
  deliverynet  NetDevs[deliver2].OnBoot          --       (true)
  deliverynet  NetDevs[deliver2].Ipaddr          --       10.0.30.250
  deliverynet  NetDevs[deliver2].Netmask         --       255.255.255.0
  deliverynet  NetDevs[deliver2].Tags[dynend]    --       10.10.30.50
  deliverynet  NetDevs[deliver2].Tags[dynstart]  --       10.10.30.10
  deliverynet  NetDevs[deliver2].Tags[network]   --       10.0.30.0
  deliverynet  NetDevs[deliver1].Type            --       (ethernet)
  deliverynet  NetDevs[deliver1].OnBoot          --       (true)
  deliverynet  NetDevs[deliver1].Ipaddr          --       10.0.20.250
  deliverynet  NetDevs[deliver1].Netmask         --       255.255.255.0
  deliverynet  NetDevs[deliver1].Primary         --       (true)
  deliverynet  NetDevs[deliver1].Tags[network]   --       10.0.20.0
  deliverynet  NetDevs[deliver1].Tags[dynend]    --       10.10.20.50
  deliverynet  NetDevs[deliver1].Tags[dynstart]  --       10.10.20.10

Now the templates of ``dhcpd`` and/or ``dnsmasq`` must be modified.

.. code-block:: shell

   wwctl overlay edit host etc/dhcpd.conf.ww
   wwctl overlay edit host etc/dnsmasq.d/ww4-hosts.ww

For the ``dhcp`` template you should add following lines

.. code-block::

   {{/* multiple networks */}}
   {{- range $node := $.AllNodes}}
   {{- if eq $node.Id.Get "deliverynet" }}
   {{- range $netname, $netdev := $node.NetDevs}}
   # network {{ $netname }}
   subnet {{$netdev.Tags.network.Get}} netmask {{$netdev.Netmask.Get}} {
       max-lease-time 120;
       range {{$netdev.Tags.dynstart.Get}} {{$netdev.Tags.dynend.Get}};
       next-server {{$netdev.Ipaddr.Get}};
   }
   {{- end }}
   {{- end }}
   {{- end }}

and for the ``dnsmasq`` the following lines should be added

.. code-block::

   {{/* multiple networks */}}
   {{- range $node := $.AllNodes}}
   {{- if eq $node.Id.Get "deliverynet" }}
   {{- range $netname, $netdev := $node.NetDevs}}
   # network {{ $netname }}
   dhcp-range={{$netdev.Tags.dynstart.Get}},{{$netdev.Tags.dynend.Get}},{{$netdev.Netmask.Get}},6h
   {{- end }}
   {{- end }}
   {{- end }}

Note that the ``{{- if eq $node.Id.Get "deliverynet" }}`` is used to identify
the dummy host which carries the network information.
