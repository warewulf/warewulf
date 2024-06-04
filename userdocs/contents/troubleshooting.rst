Troubleshooting
===============

iPXE
----

If you're using iPXE to boot (the default), you can get a command
prompt by pressing with C-b during boot.

From there, you can use the same commands from default.ipxe to 

- at iPXE start, C-b to get a command prompt
- ifconf
- set uri http://192.168.3.11:9873/provision/${mac}
- imgextract --name container ${uri}?stage=container&compress=gz
- imgextract --name system ${uri}?stage=system&compress=gz
- imgextract --name runtime ${uri}?stage=runtime&compress=gz
- boot ${uri}?stage=kernel initrd=container initrd=system initrd=runtime debug rdinit=/bin/sh

GRUB
----

If you're using GRUB to boot, you can get a command prompt by pressing
"c" when prompted.

From the GRUB command prompt, you can enter the same commands that you
would otherwise find in `/etc/warewulf/grub/grub.cfg.ww`.

For example, the following commands perform a (relatively) normal
Warewulf boot. (Substitute your Warewulf server's IP address in place
of 10.0.0.1, and update the port number if you have changed it from
the default of 9873.)

.. code-block::

   uri="(http,10.0.0.1:9873)/provision/${net_default_mac}?assetkey="
   linux "${uri}&stage=kernel" wwid=${net_default_mac}
   initrd "${uri}&stage=container&compress=gz" "${uri}&stage=system&compress=gz" "${uri}&stage=runtime&compress=gz"
   boot

- The ``uri`` variable points to ``warewulfd`` for future
  reference. ``${net_default_mac}`` provides Warewulf with the MAC
  address of the booting node, so that Warewulf knows what container
  and overlays to provide it.

- The ``linux`` command tells GRUB what kernel to boot, as provided by
  ``warewulfd``. The ``wwid`` kernel argument helps ``wwclient``
  identify the node during runtime.

- The ``initrd`` command tells GRUB what images to load into memory for
  boot. In a typical environment this is used to load a minimal
  "initial ramdisk" which, then, boots the rest of the
  system. Warewulf, by default, loads the entire image as an initial
  ramdisk, and also loads the system and runtime overlays at this time
  time.

- The ``boot`` command tells GRUB to boot the system with the
  previously-defined configuration.

.. note::

   This example does not provide ``assetkey`` information to
   ``warewulfd``. If your nodes have defined asset tags, provide it in
   the ``uri`` variable for the node you are trying to boot.

For example, you may want to try booting to a pre-init shell with
debug logging enabled. To do so, substitute the ``linux`` command
above.

.. code-block::

   linux "${uri}&stage=kernel" wwid=${net_default_mac} debug rdinit=/bin/sh

.. note::

   You may be more familiar with specifying ``init=`` on the kernel
   command line. ``rdinit`` indicates "ramdisk init." Since Warewulf,
   by default, boots the node image as an initial ramdisk, we must use
   ``rdinit=`` here.

warewulfd
---------

- curl -L http://192.168.3.10:9873/efiboot/grub.cfg
- http://192.168.3.10:9873/ipxe/E6:92:39:49:7B:03
- http://192.168.3.11:9873/provision/e6:92:39:49:7b:03?assetkey=${asset}&uuid=${uuid}
  - ${uri_base}&stage=kernel
  - ${uri_base}&stage=container&compress=gz
  - ${uri_base}&stage=system&compress=gz
  - ${uri_base}&stage=runtime&compress=gz
http://192.168.3.11:9873/provision/e6:92:39:49:7b:03?stage=grub
http://192.168.3.11:9873/provision/e6:92:39:49:7b:03?stage=shim

tftp
----

- warewulf/grub.cfg
  - conf="(http,192.168.3.11:9873)/efiboot/grub.cfg"
  - shim="(http,192.168.3.11:9873)/efiboot/shim.efi"
- warewulf/grub.efi
- warewulf/grubx64.efi
- warewulf/shim.efi
