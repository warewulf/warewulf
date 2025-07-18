=============
Using dnsmasq
=============

``dnsmasq`` is the default  ``dhcpd`` and ``tftp`` service. This can be configured
in the server configuration file, typically at ``/etc/warewulf/warewulf.conf``:

.. code-block:: yaml

   tftp:
     systemd name: dnsmasq
   dhcp:
     systemd name: dnsmasq

The configuration of ``dnsmasq`` often doesn't need to be changed, as the
default configuration includes all files with following pattern
``/etc/dnsmasq.d/*conf`` into its configuration. This configuration is created
by the overlay template ``host:/etc/dnsmasq.d/ww4-hosts.conf.ww``.

.. note::

   In certain distributions, such as Rocky Linux 9, ``dnsmasq`` is configured to
   listen locally via the ``interface=lo`` option by default. Replace this entry
   in ``/etc/dnsmasq.conf`` with the interface associated with your Warewulf
   network, or remove/comment out the interface option entirely to enable
   listening on all interfaces.

Once the Warewulf configuration has been updated, re-deploy the configuration
and restart ``warewulfd``.

.. code-block:: shell

   wwctl configure --all
   systemctl restart warewulfd.service
