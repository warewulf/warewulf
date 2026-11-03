=============
Using dnsmasq
=============

As an experimental feature, it is possible to use ``dnsmasq`` instead of the ISC
``dhcpd`` server and ``TFTP`` server.

In order to keep the configuration manageable, the file ``/etc/dnsmasq.d/ww4-hosts.conf`` is created and must
be included in the main ``dnsmasq.conf`` via the ``conf-dir=/etc/dnsmasq.d``
option.

Installation
============

Before the installation, make sure that ``dhcpd`` and ``tftp`` are disabled.
You can do that with the commands:

.. code-block:: shell

   systemctl disable --now dhcpd
   systemctl disable --now tftp

Now you can install ``dnsmasq``.

.. code-block:: shell

   # Rocky Linux
   dnf install dnsmasq

   # SUSE
   zypper install dnsmasq

After the installation, instruct ``warewulf`` to use ``dnsmasq`` as its
``dhcpd`` and ``tftp`` service. This is done in the server configuration file,
typically at ``/etc/warewulf/warewulf.conf``:

.. code-block:: yaml

   tftp:
     systemd name: dnsmasq
     overlays: dnsmasq,tftproot
   dhcp:
     systemd name: dnsmasq
     overlays: dnsmasq

The configuration of ``dnsmasq`` often doesn't need to be changed, as the
default configuration includes all files with following pattern
``/etc/dnsmasq.d/*conf`` into its configuration. This configuration is created
by the overlay templates ``dnsmasq:/etc/dnsmasq.d/ww4-hosts.conf.ww`` and
``dnsmasq:/etc/dnsmasq.d/ww4-listen.conf.ww``.

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

Router Advertisements
======================

When IPv6 is configured, the generated ``ww4-hosts.conf`` enables dnsmasq's
Router Advertisements (``enable-ra``): dnsmasq advertises the provisioning
prefix and acts as the default router. On a network where an upstream router
already sends Router Advertisements, two RA sources on the same segment can
lead to unexpected client addressing and routing. The ``dhcp:router
advertisements`` setting in ``warewulf.conf`` selects where RAs come from (it
has no effect with the ISC ``dhcpd`` server, which never sends RAs):

- ``warewulf`` (default): dnsmasq sends RAs and is the default router.
- ``warewulf-no-default-route``: dnsmasq sends RAs with a router lifetime of 0
  (``ra-param=*,0,0``), so clients do not use the Warewulf server as their
  default gateway, while the RAs still carry the M or O flag that iPXE needs
  before it starts DHCPv6.
- ``upstream``: Warewulf sends no RAs. The upstream router must set the M flag
  (when ``range6 start``/``range6 end`` are configured) or the O flag
  (otherwise) in its RAs, or iPXE never requests DHCPv6 and boot fails with
  ``Nothing to boot: No such file or directory``.

See :ref:`the configuration reference <server-configuration>` for details.
