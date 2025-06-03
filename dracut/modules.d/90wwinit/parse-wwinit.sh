#!/bin/sh
# root=wwinit|wwinit:*

[ -z "$root" ] && root=$(getarg root=)

case "${root}" in
wwinit|wwinit:*)
    info "warewulf: root=${root}"

    export wwinit_uri="$(getarg wwinit.uri)"
    if [ -n "${wwinit_uri}" ]; then
        info "warewulf: Found root=${root} and wwinit.uri=${wwinit_uri}. Will boot from Warewulf."
        rootok=1
    else
        die "warewulf: Found root=${root} but no wwinit.uri. Cannot boot from Warewulf."
    fi

    export wwinit_uuid=$(dmidecode -s system-uuid)
    export wwinit_assetkey=$(dmidecode -s chassis-asset-tag)

    wwinit_tmpfs_size="$(getarg wwinit.tmpfs.size)"
    if [ -n "$wwinit_tmpfs_size" ]; then
        export wwinit_tmpfs_size_option="-o size=${wwinit_tmpfs_size}"
    fi

    case "${root}" in
    wwinit)
        export wwinit_root_device="tmpfs"
        ;;
    wwinit:*)
        export wwinit_root_device="${root#wwinit:}"
        ;;
    esac

    case "${wwinit_root_device}" in
    initramfs|rootfs)
        info "warewulf: using tmpfs in stead of ${wwinit_root_device}"
        export wwinit_root_device=tmpfs
        ;;
    esac
    ;;
esac
