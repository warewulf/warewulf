#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]
then
    info "root=${root}"
    export wwinit_uri="http://$(getarg wwinit.uri)/provision/$(getarg wwid)?assetkey=$(dmidecode -s chassis-asset-tag)&uuid=$(dmidecode -s system-uuid)&stage="
    export wwinit_container="${wwinit_uri}container&compress=gz"; info "wwinit.container=${wwinit_container}"
    export wwinit_system="${wwinit_uri}system&compress=gz"; info "wwinit.system=${wwinit_system}"
    export wwinit_runtime="${wwinit_uri}runtime&compress=gz"; info "wwinit.runtime=${wwinit_runtime}"
    wwinit_kmods_passed=$(getarg wwinit.kmods=)
    [ -n "$wwinit_kmods_passed" ] && export wwinit_kmods="${wwinit_uri}kmods&compress=gz"; info "wwinit.kmods=${wwinit_kmods}"

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
