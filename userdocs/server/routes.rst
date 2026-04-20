.. _server-routes:

=============
Server Routes
=============

The Warewulf provisioning daemon, ``warewulfd``, serves all boot and provisioning
resources over HTTP. Each resource type has a dedicated route, and nodes are
identified by their Warewulf ID (``wwid``).

``{wwid}`` is typically the node's default MAC address: a colon-separated
hexadecimal string, e.g., ``aa:bb:cc:dd:ee:ff``. Dashes are accepted in place
of colons and are normalized automatically.

.. note::

   The port ``warewulfd`` listens on is configured with ``warewulf:port`` in
   ``warewulf.conf`` (default: ``9873``). When TLS is enabled, a second listener
   is started on the port configured with ``warewulf:tls port`` (default: ``9874``).

URL Patterns
============

Every provisioning route (except ``/overlay-file/`` and ``/status``) supports
six equivalent URL patterns for specifying the node identity:

.. code-block:: none

   /{stage}/{wwid}           # wwid in path
   /{stage}?wwid={wwid}      # wwid as query parameter
   /{stage}                  # wwid resolved from ARP cache

   /provision/{wwid}?stage={stage}
   /provision?wwid={wwid}&stage={stage}
   /provision?stage={stage}  # wwid resolved from ARP cache

The ``/efiboot/`` route is an exception: the path segment contains the boot
file name rather than a wwid, and the node is identified via ``?wwid=`` or ARP:

.. code-block:: none

   /efiboot/{file}                    # file in path, wwid from ARP
   /efiboot/{file}?wwid={wwid}        # file in path, explicit wwid
   /efiboot?wwid={wwid}&file={file}   # both as query parameters

Common Query Parameters
=======================

Most provisioning routes accept the following query parameters:

* ``wwid``: Warewulf ID of the node, typically its default MAC address. Used
  when the node identity is not embedded in the URL path. Takes priority over
  ARP lookup; the path always takes priority over the ``wwid`` query parameter.

* ``assetkey``: Hardware asset tag. If the node has an ``AssetKey`` configured
  in ``nodes.conf``, the server requires this parameter to match before serving
  any content. See :ref:`Security <server-routes-security>` below.

* ``uuid``: System UUID of the requesting node. Accepted for logging purposes.

* ``compress``: Compression format for the response. The only supported value
  is ``gz``. When ``compress=gz`` is specified, the server serves a pre-built
  gzip-compressed version of the file. If no compressed version exists, the
  server returns ``404 Not Found``.

Provisioning Routes
===================

``/ipxe/{wwid}``
----------------

Serves a rendered iPXE boot script for the node identified by ``{wwid}``.

The script is rendered as a Go template from a file in
``/etc/warewulf/ipxe/``. The specific template used is determined by the
node's ``Ipxe`` field (defaulting to ``default``); for example, a node with
``Ipxe: dracut`` receives the template from
``/etc/warewulf/ipxe/dracut.ipxe``.

If the requesting node is not known to Warewulf, the server falls back to
serving ``/etc/warewulf/ipxe/unconfigured.ipxe``.

**Query parameters:** ``assetkey``, ``uuid``

``/kernel/{wwid}``
------------------

Serves the raw kernel binary for the node identified by ``{wwid}``. The
kernel is taken from the node's assigned image.

**Query parameters:** ``assetkey``, ``uuid``, ``compress``

``/image/{wwid}``
-----------------

Serves the raw OS image file for the node identified by ``{wwid}``.

**Query parameters:** ``assetkey``, ``uuid``, ``compress``

``/initramfs/{wwid}``
---------------------

Serves the initramfs binary for the node identified by ``{wwid}``. The
initramfs is extracted from the node's assigned image based on the node's
kernel version. This route is used in two-stage boot configurations. See
:ref:`booting with dracut` for details.

**Query parameters:** ``assetkey``, ``uuid``, ``compress``

``/system/{wwid}``
------------------

Serves the system overlay image for the node identified by ``{wwid}``.
The system overlay is rendered at provisioning time and contains
configuration files that are static for the lifetime of the boot.

When ``autobuild overlays`` is enabled in ``warewulf.conf``, the server
will automatically rebuild the overlay if it is out of date relative to
``nodes.conf`` or the overlay source files.

**Query parameters:** ``assetkey``, ``uuid``, ``compress``, ``overlay``

* ``overlay``: A comma-separated list of overlay names. When specified, only
  the named overlays are served (rather than the node's full system overlay
  set).

``/runtime/{wwid}``
-------------------

Serves the runtime overlay image for the node identified by ``{wwid}``.
The runtime overlay is rendered on demand and may contain node-specific
secrets. ``wwclient`` fetches the runtime overlay periodically during normal
operation.

When ``warewulf:secure`` is enabled in ``warewulf.conf``, this route requires
that the request originate from a privileged TCP port (port number less than
1024). This prevents unprivileged users on a node from retrieving the runtime
overlay.

When TLS is enabled in ``warewulf.conf``, this route requires that the request
arrive over HTTPS. Plain-HTTP requests are rejected with ``403 Forbidden``. The
HTTPS listener port is configured with ``warewulf:tls port``.

**Query parameters:** ``assetkey``, ``uuid``, ``compress``, ``overlay``

* ``overlay``: A comma-separated list of overlay names. Same behavior as for
  ``/system/``.

``/overlay-file/{overlay}/{path}``
----------------------------------

Provides direct access to an individual file within a named overlay. This
route uses a different URL structure than the other provisioning routes: the
overlay name is in the second path segment, and the file path within the overlay
follows.

If the ``render`` parameter is provided, the file is rendered as a Go template
for the specified node and the rendered content is returned. If ``render`` is
absent, the raw file bytes are returned without any template processing.

If the requested path does not end in ``.ww`` but a ``.ww``-suffixed version of
the file exists, and a ``render`` node is specified, the server automatically
serves the ``.ww`` template.

**Query parameters:**

* ``render``: Node ID to render the template for. If not specified, the raw
  file is returned.

.. note::

   This route does not require authentication via ``assetkey`` and does not
   perform node lookup by hardware address.

``/tpm-quote/``
---------------

Receives and stores a TPM quote and event log uploaded by a node during boot.
This route allows a cluster node to securely authenticate itself and verify its boot
integrity. The quote is verified during the subsequent challenge request.

**Query parameters:** ``wwid``

``/tpm-challenge``
------------------

Generates and serves a Credential Activation Challenge for the node identified
by ``wwid``. The server first verifies the node's previously uploaded TPM quote,
event log, and GRUB measurements. If verification succeeds, the server returns an
encrypted challenge that only the node's specific TPM can decrypt. This challenge
is used by the node to securely unlock its runtime overlay.

**Query parameters:** ``wwid``

``/efiboot/{file}``
-------------------

Serves EFI boot files for GRUB-based booting. The requesting node is
identified by ``?wwid=`` (preferred) or, if not supplied, by an ARP lookup
of the client's IP address against the kernel's ARP cache (``/proc/net/arp``).
This route is intended for EFI HTTP Boot clients, where the firmware fetches
a boot URI from DHCP and cannot perform variable substitution.

.. note::

   On large clusters the kernel's default ARP cache limits may be exceeded,
   causing node identification to fail. See
   :ref:`arp-cache-overflow-on-large-clusters` for tuning guidance.

The ``{file}`` component determines what is served:

* ``shim.efi``: Serves the ``shim.efi`` binary extracted from the node's
  assigned image.
* ``grub.efi`` (or ``grubx64.efi``, ``grubaa64.efi``, ``grubia32.efi``,
  ``grubarm.efi``, ``grub-tpm.efi``): Serves the GRUB EFI binary extracted
  from the node's assigned image.
* ``grub.cfg``: Serves a rendered GRUB configuration file from
  ``/etc/warewulf/grub/grub.cfg.ww``. The configuration is rendered as a Go
  template for the identified node.

Because ``shim.efi`` resolves subsequent files relative to its own load URL,
GRUB and ``grub.cfg`` are also fetched from the ``/efiboot/`` path. The
``grub.cfg`` served by this route uses ``${net_default_mac}`` to embed the
node's wwid in all further provisioning URLs, directing subsequent requests
to the per-node ``/grub/{wwid}`` route.

**Query parameters:** ``assetkey``, ``uuid``, ``wwid``, ``file``

* ``file``: The EFI file to serve (``shim.efi``, ``grub.efi``, or ``grub.cfg``).
  Used when the file name is not embedded in the URL path.

.. note::

   ``/efiboot/`` is the recommended route for EFI HTTP Boot clients. For
   TFTP-booted GRUB clients that know their own wwid, use
   ``/grub/{wwid}`` to fetch the per-node GRUB configuration directly.

``/grub/{wwid}``
----------------

Serves a rendered GRUB configuration file for the node identified by
``{wwid}``. The configuration is rendered from
``/etc/warewulf/grub/grub.cfg.ww`` as a Go template. This route is the
preferred method for TFTP-booted GRUB clients to fetch their per-node
configuration, as the node identity is explicit in the URL rather than
resolved via ARP.

**Query parameters:** ``assetkey``, ``uuid``, ``wwid``

``/provision/{wwid}``
---------------------

.. deprecated::

   This route is maintained for backwards compatibility. New configurations
   should use the dedicated routes described above.

A legacy dispatcher route. The provisioning stage is determined by the
``stage`` query parameter, which is dispatched to the appropriate handler:

* ``stage=ipxe`` → ``/ipxe/``
* ``stage=kernel`` → ``/kernel/``
* ``stage=image`` → ``/image/``
* ``stage=initramfs`` → ``/initramfs/``
* ``stage=system`` → ``/system/``
* ``stage=runtime`` → ``/runtime/``
* ``stage=grub`` → ``/grub/``

**Query parameters:** ``stage`` (required), ``assetkey``, ``uuid``, ``compress``,
``overlay``

Status Route
============

``/status``
-----------

Returns a JSON object containing the last-known provisioning status for all
nodes that have contacted the server. No authentication is required.

.. code-block:: none

   {
     "nodes": {
       "node01": {
         "node name": "node01",
         "stage": "RUNTIME_OVERLAY",
         "sent": "2 kB",
         "ipaddr": "10.0.1.1",
         "last seen": 1712345678
       }
     }
   }

The ``stage`` field reflects the most recent provisioning stage completed for
the node. Possible values include ``IPXE``, ``KERNEL``, ``IMAGE``,
``INITRAMFS``, ``SYSTEM_OVERLAY``, ``RUNTIME_OVERLAY``, and ``EFI``.

REST API
========

When enabled in ``warewulf.conf``, ``warewulfd`` exposes a REST API under
``/api/``. The API provides programmatic access to nodes, profiles, images,
and overlays. Interactive documentation is available at ``/api/docs``.

When TLS is enabled, access to the REST API can additionally be restricted to
HTTPS-only requests by setting ``api: tls: true`` in ``warewulf.conf``.

See :ref:`rest-api` for full details.

.. _server-routes-security:

Security
========

Several mechanisms are available to restrict access to provisioning routes.

Asset key validation
--------------------

If a node is configured with an ``AssetKey`` in ``nodes.conf``, the Warewulf
server will only respond to provisioning requests that include a matching
``?assetkey=`` query parameter. Requests with a missing or incorrect asset key
receive ``401 Unauthorized``. The asset key is typically a hardware-level
firmware string (an "asset tag") that is accessible only with root or physical
access.

.. code-block:: console

   # wwctl node set node01 --assetkey "SYSTEM-ASSET-TAG"

Secure mode
-----------

When ``warewulf:secure`` is set to ``true`` in ``warewulf.conf``, the
``/runtime/`` route requires that requests originate from a privileged
TCP source port (port number less than 1024). Because only processes running as
``root`` can bind to privileged ports, this prevents unprivileged users on a
cluster node from downloading the runtime overlay.

TLS
---

When TLS is enabled in ``warewulf.conf``, the ``/runtime/`` route
rejects plain-HTTP requests with ``403 Forbidden``. Runtime overlays must be
fetched over HTTPS. Because iPXE and GRUB cannot handle HTTPS, the kernel,
image, and system overlay continue to be served over plain HTTP even when TLS
is enabled.

See :ref:`Security <server/security>` for instructions on enabling TLS and
generating certificates.
