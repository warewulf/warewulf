#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]
then
    info "root=${root}"
    export wwinit_uuid=$(dmidecode -s system-uuid)
    export wwinit_assetkey=$(dmidecode -s chassis-asset-tag)
    export wwinit_uri="$(getarg wwinit.uri)"

    wwinit_tmpfs_size=$(getarg wwinit.tmpfs.size=)
    if [ -n "$wwinit_tmpfs_size" ]
    then
        info "wwinit.tmpfs.size=${wwinit_tmpfs_size}"
        export wwinit_tmpfs_size_option="-o size=${wwinit_tmpfs_size}"
    fi

    if [ -n "${wwinit_uri}" ]
    then
        info "Found root=${root} and a Warewulf server uri. Will boot from Warewulf."
        rootok=1
    else
        die "Found root=${root} but no Warewulf server uri. Cannot boot from Warewulf."
    fi
fi
