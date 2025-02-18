=============
Using dnsmasq
=============

As an experimental feature, it is possible to use ``dnsmasq`` instead of the ISC
``dhcpd`` server and ``TFTP`` server.

In order to keep the file ``/etc/dnsmasq.d/ww4-hosts.conf`` is created and must
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
