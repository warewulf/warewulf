==========
Templating
==========

Warewulf uses the ``text/template`` engine to convert dynamic content into static content and auto-populate files with the appropriate data on demand.

In Warewulf, you can find templates both for the provisioning services (e.g. ``/etc/warewulf/ipxe/``, ``/etc/warewulf/dhcp/``, and ``/etc/warewulf/hosts.tmpl``) as well as within the runtime and system overlays.

(more documentation coming soon)

Examples
========

range
-----

iterate over elements of an array

.. code-block:: go

   {{ range $devname, $netdev := .NetDevs }}
       # netdev = {{ $netdev.Hwaddr }}
   {{ end }}

increment variable in loop
^^^^^^^^^^^^^^^^^^^^^^^^^^

iterate over elements of an array and increment ``i`` each loop cycle

.. code-block:: go

   {{ $i := 0 }}
   {{ range $devname, $netdev := .NetDevs }}
       # netdev{{$i}} = {{ $netdev.Hwaddr }}
       {{ $i = inc $i }}
   {{ end }}

decrement
^^^^^^^^^

iterate over elements of an array and decrement ``i`` each loop cycle

.. code-block:: go

   {{ $i := 10 }}
   {{ range $devname, $netdev := .NetDevs }}
       # netdev{{$i}} = {{ $netdev.Hwaddr }}
       {{ $i = dec $i }}
   {{ end }}