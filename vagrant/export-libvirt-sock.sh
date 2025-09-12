#!/bin/sh

pidfile=".export-libvirt-sock.pid"

pid=$(cat $pidfile 2>/dev/null || true)

if [ "$pid" != "" ]; then
	if grep "/var/tmp/libvirt.sock:" /proc/$pid/cmdline >/dev/null 2>&1; then
		exit 0
	fi
fi

nohup vagrant ssh wwctl -- -o ServerAliveInterval=5 -o ServerAliveCountMax=1 -f -R /var/tmp/libvirt.sock:/var/run/libvirt/libvirt-sock -N >/dev/null 2>&1 &
echo $! > .export-libvirt-sock.pid
