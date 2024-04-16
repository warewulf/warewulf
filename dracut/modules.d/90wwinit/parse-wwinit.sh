#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]
then
    info "root=${root}"
    wwinit_image=$(getarg wwinit.image=); debug "wwinit.image=${wwinit_image}"
    wwinit_system=$(getarg wwinit.system=); debug "wwinit.system=${wwinit_system}"
    wwinit_runtime=$(getarg wwinit.runtime=); debug "wwinit.runtime=${wwinit_runtime}"
    wwinit_kmods=$(getarg wwinit.kmods=); debug "wwinit.kmods=${wwinit_kmods}"

    wwinit_tmpfs_size=$(getarg wwinit.tmpfs.size=)
    if [ -n "$wwinit_tmpfs_size" ]
    then
        debug "wwinit.tmpfs.size=${wwinit_tmpfs_size}"
        wwinit_tmpfs_size_option="-o size=${wwinit_tmpfs_size}"
    fi

    if [ -n "${wwinit_image}" ]
    then
        info "Found root=${root} and a Warewulf container image. Will boot from Warewulf."
        rootok=1
    else
        die "Found root=${root} but no container image. Cannot boot from Warewulf."
    fi
fi
