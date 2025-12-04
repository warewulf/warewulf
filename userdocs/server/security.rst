========
Security
========

While certain parallelization and high performance library capabilities still
require lowering the security threshold within a cluster, Warewulf strives to
support good security practices within the cluster wherever possible.

Provisioning Security
=====================

Provisioning is, by default, a relatively "insecure" process: there is generally
nothing preventing a user on a cluster node from spoofing a provision request
and downloading the node image and overlays for inspection. If any of these
include secrets (e.g., private keys) they are at risk of exposure.

There are multiple ways to secure the Warewulf provisioning process:

* The best way to secure the provisioning process is to dedicate a vLAN
  specifically for provisioning, and then not make that vLAN available in the
  provisioned environment. Warewulf can be used in such an environment (without
  ``wwclient``) but you must consult your switch documentation and features to
  implement a default vLAN for provisioning and to ensure that the runtime
  operating system is configured for a different tagged vLAN once booted.

* Warewulf can leverage hardware "asset tags" which almost all vendors support.
  This is a configurable firmware string that is accessible only via root or
  physical access. During provisioning (as well as post provisioning via
  ``wwclient``) Warewulf sends the detected asset tag to the Warewulf server as
  a "shared secret" token. If the node is also configured with an ``asset key``
  on the Warewulf server (e.g., via ``wwctl node set --assetkey "..."``), the
  Warewulf server will only respond to requests with a matching asset tag.

* If the Warewulf server is configured with ``warewulf:secure: true``, then it
  will only provide the runtime overlay to a ``wwclient`` communicating from a
  privileged (< 1024) TCP port. This prevents unprivileged cluster users from
  being able to retrieve the runtime overlay.

* When the nodes are booted via `shim` and `grub` Secure Boot can be enabled.
  This means that the nodes only boot the kernel which is provided by the
  distributor and also custom complied modules can't be loaded.

* Tls (transport layer security) can be enabled for the warewulf server. However
  the https is only used for runtime overlays. The kernel and system image are *always*
  transfered unencrypted as iPXE and grub can't handle https.

SELinux
=======

The Warewulf server can be run with SELinux enabled in "targeted" and
"enforcing" mode.

For more information about running SELinux-enabled cluster node images, see
:ref:`SELinux-Enabled Images <selinux_images>`.

firewalld
=========

If the Warewulf server is running ``firewalld``, the following services must be
added for them to function:

.. code-block:: console

   firewall-cmd --permanent --add-service=warewulf
   firewall-cmd --permanent --add-service=dhcp
   firewall-cmd --permanent --add-service=nfs
   firewall-cmd --permanent --add-service=tftp
   firewall-cmd --reload

.. note::

   The DHCP, TFTP, and NFS services may be managed manually, apart from the
   Warewulf server. In that case, they may be omitted from the ``firewalld``
   configuration on the Warewulf server; but they must be accessible from where
   they are served.

nftables
========

If the Warewulf server is running ``nftables`` directly, without ``firewalld``,
ensure that TCP port ``9873`` must be permitted for cluster nodes to communicate
with the Warewulf server.

.. code-block:: console

   nft add rule inet filter input tcp dport 9873 accept
   nft list ruleset >/etc/nftables.conf
   systemctl restart nftables


HTTPS
=====

The https functionality can be enabled by setting

.. code-block:: yaml
..
.. warewulf:
..   tls:         true
..   secure port: 9874
..

Which will enable a https server on the secure port. The certificate and key
can be created as self signed key with ``wwctl configure keys --create``. The
keys and certificate are stored if not configured otherwise as

.. code-block:: console
..
.. /etc/warewulf/keys/warewulf.crt # PEM certificate
.. /etc/warewulf/keys/warewulf.key # PEM RSA Key



For the key and certifcate generation no addiotional parameters can be set, but
you can import your own keys, with ``wwctl configure keys import``.

If HTTPS is enabled the delivery of the runtime overlays is disabled over HTTP, so
you **must** use `wwclient` to get the runtime overlays.

