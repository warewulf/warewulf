#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]
then
    info "root=${root}"
    wwinit_uri="$(getarg wwinit.uri)"
    export wwinit_container="${wwinit_uri}&stage=container&compress=gz"; info "wwinit_container=${wwinit_container}"
    export wwinit_system="${wwinit_uri}&stage=system&compress=gz"; info "wwinit_system=${wwinit_system}"
    export wwinit_runtime="${wwinit_uri}&stage=runtime&compress=gz"; info "wwinit_runtime=${wwinit_runtime}"
    if [ -n "$(getarg wwinit.KernelOverride)" ]
    then
        export wwinit_kmods="${wwinit_uri}&stage=kmods&compress=gz"; info "wwinit_kmods=${wwinit_kmods}"
    fi

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
