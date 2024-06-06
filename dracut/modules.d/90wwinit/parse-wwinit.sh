#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]
then
    info "root=${root}"
    export wwinit_container=$(getarg wwinit.container=); info "wwinit.container=${wwinit_container}"
    export wwinit_system=$(getarg wwinit.system=); info "wwinit.system=${wwinit_system}"
    export wwinit_runtime=$(getarg wwinit.runtime=); info "wwinit.runtime=${wwinit_runtime}"
    export wwinit_kmods=$(getarg wwinit.kmods=); info "wwinit.kmods=${wwinit_kmods}"

    wwinit_tmpfs_size=$(getarg wwinit.tmpfs.size=)
    if [ -n "$wwinit_tmpfs_size" ]
    then
        info "wwinit.tmpfs.size=${wwinit_tmpfs_size}"
        export wwinit_tmpfs_size_option="-o size=${wwinit_tmpfs_size}"
    fi

    if [ -n "${wwinit_container}" ]
    then
        info "Found root=${root} and a Warewulf container image. Will boot from Warewulf."
        rootok=1
    else
        die "Found root=${root} but no container image. Cannot boot from Warewulf."
    fi
fi
