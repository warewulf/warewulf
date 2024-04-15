#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]; then
    info "Found root=${root}: will boot from Warewulf."
    rootok=1
fi