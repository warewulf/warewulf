#!/bin/sh
# root=wwinit

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ] || [ "${root}" = "persistent" ] ; then
    info "warewulf: root=${root}"
    export wwinit_uuid=$(dmidecode -s system-uuid)
    export wwinit_assetkey=$(dmidecode -s chassis-asset-tag)
    export wwinit_uri="$(getarg wwinit.uri)"
    export wwinit_ignition="$(getarg wwinit.ignition)"
    export wwinit_ip="$(getarg wwinit.ip)"
    export wwinit_imagename="$(getarg wwinit.imagename)"
    export wwinit_id=$(getarg wwinit.id)
    wwinit_tmpfs_size=$(getarg wwinit.tmpfs.size=)
    if [ -n "$wwinit_tmpfs_size" ] ; then
        info "warewulf: wwinit.tmpfs.size=${wwinit_tmpfs_size}"
        export wwinit_tmpfs_size_option="-o size=${wwinit_tmpfs_size}"
    fi
    if [ "${root}" = "persistent" ] ; then 
        export wwinit_persistent="1"; info "warewulf: wwinit_persistent=$wwinit_persistent"
    fi
    if [ -n "${wwinit_uri}" ] ; then
        info "warewulf: Found root=${root} and a Warewulf server uri. Will boot from Warewulf."
        rootok=1
    else
        die "warewulf: Found root=${root} but no Warewulf server uri. Cannot boot from Warewulf."
    fi
fi
