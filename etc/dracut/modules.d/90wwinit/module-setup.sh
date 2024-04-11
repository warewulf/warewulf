#!/bin/bash

check() {
    # Don't include in hostonly mode
    [[ $hostonly ]] && return 1

    # Don't include by default
    return 255
}

depends() {
    echo network
    return 0
}

install() {
    inst_hook cmdline 30 "$moddir/parse-root.sh"
    inst_hook mount 30 "$moddir/load-root.sh"
    inst_hook pre-pivot 99 "$moddir/fix-selinux.sh"
}
