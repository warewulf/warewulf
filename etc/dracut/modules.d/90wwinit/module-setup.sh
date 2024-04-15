#!/bin/bash

check() {
    # Don't include in hostonly mode
    [[ $hostonly ]] && return 1

    # Don't include by default
    return 255
}

install() {
    inst_multiple cpio curl
    inst_hook cmdline 30 "$moddir/parse-root.sh"
    inst_hook pre-mount 30 "$moddir/load-root.sh"
}
