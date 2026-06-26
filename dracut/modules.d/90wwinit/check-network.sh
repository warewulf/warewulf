#!/bin/sh
# initqueue/finished hook: when booting from Warewulf, hold the initqueue
# open until the network has a usable global-scope address (IPv4 or IPv6)
# so that the pre-mount load-wwinit.sh hook can reach the Warewulf server.
# Configurations using NetworkManager wait via nm-wait-online instead (see
# module-setup.sh); this hook covers systemd-networkd and other network
# configurations that do not provide their own readiness gate.
[ -z "$wwinit_root_device" ] && exit 0
# Ignore IPv6 addresses still undergoing (tentative) or having failed
# (dadfailed) duplicate address detection: they are not yet usable.
ip -o addr show scope global 2>/dev/null \
    | grep -v -e tentative -e dadfailed \
    | grep -q inet
