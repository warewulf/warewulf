=======
Dnsmasq
=======

Usage
=====

As experimental feature its possible to use `dnsmasq` instead of the ISC `dhcpd` server in combination 
with a `tFTP` server. The `dnsmasq` service is then acting as `dhcp` and `tftp` server. In order to keep 
the file `/etc/dnsmasq.d/ww4-hosts.conf` is created and must be included in the main `dnsmasq.conf` via 
the `conf-dir=/etc/dnsmasq.d` option.

Addionally in the configuration file `warewulf.conf` in the sections `dhcp` and `tftp` the systemd name of 
dnsmasq must set for the option `systemd name`.

After this configuration steps its recommended to rebuild the host overlay with `wwctl overlay build -H` and
the the services should be configured with `wwctl configure -a`.
