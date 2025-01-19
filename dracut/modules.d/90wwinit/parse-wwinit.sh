#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]
then
    info "root=${root}"
    uuid=$(dmidecode -s system-uuid)
    assetkey=$(dmidecode -s chassis-asset-tag | sed -E -e 's/(^ +| +$)//g' -e 's/^(Unknown|Not Specified)$//g' -e 's/ /_/g')
    wwinit_uri="$(getarg wwinit.uri)?assetkey=${assetkey}&uuid=${uuid}"
    export wwinit_image="${wwinit_uri}&stage=image&compress=gz"; info "wwinit_image=${wwinit_image}"
    export wwinit_system="${wwinit_uri}&stage=system&compress=gz"; info "wwinit_system=${wwinit_system}"
    export wwinit_runtime="${wwinit_uri}&stage=runtime&compress=gz"; info "wwinit_runtime=${wwinit_runtime}"

    wwinit_tmpfs_size=$(getarg wwinit.tmpfs.size=)
    if [ -n "$wwinit_tmpfs_size" ]
    then
        info "wwinit.tmpfs.size=${wwinit_tmpfs_size}"
        export wwinit_tmpfs_size_option="-o size=${wwinit_tmpfs_size}"
    fi

    if [ -n "${wwinit_image}" ]
    then
        info "Found root=${root} and a Warewulf image. Will boot from Warewulf."
        rootok=1
    else
        die "Found root=${root} but no image. Cannot boot from Warewulf."
    fi
fi
