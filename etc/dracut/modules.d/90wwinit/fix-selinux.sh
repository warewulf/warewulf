#!/bin/sh

fix_selinux()
{
    # If SELinux is disabled exit now
    getarg "selinux=0" > /dev/null && return 0

    SELINUX="enforcing"
    [ -e "$NEWROOT/etc/selinux/config" ] && . "$NEWROOT/etc/selinux/config"

    # Attempt to load SELinux Policy and fix contexts in tmpfs-root
    if [ -x "$NEWROOT/usr/sbin/load_policy" -a -x "$NEWROOT/usr/sbin/restorecon" ]; then
        local ret=0
        local out
        info "Fixing SELinux context in tmpfs-root"
        mount -o bind /sys $NEWROOT/sys
        # load_policy mounts /proc and /sys/fs/selinux in
        # libselinux,selinux_init_load_policy()
        out=$(LANG=C chroot "$NEWROOT" /usr/sbin/load_policy -i 2>&1)
        info $out

        out=$(LANG=C chroot "$NEWROOT" /usr/sbin/restorecon -F -R -e /proc -e /sys -e /dev / 2>&1)
        info $out

        umount $NEWROOT/sys/fs/selinux
        umount $NEWROOT/sys
    fi
}

fix_selinux
