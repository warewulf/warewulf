#!/bin/bash

info "Mounting tmpfs at $NEWROOT"
mount -t tmpfs ${wwinit_tmpfs_size_option} tmpfs "$NEWROOT"

for archive in "${wwinit_image}" "${wwinit_kmods}" "${wwinit_system}" "${wwinit_runtime}"
do
    if [ -n "${archive}" ]
    then
        info "Loading ${archive}"
        curl --silent -L "${archive}" | gzip -d | cpio -im --directory="${NEWROOT}"
    fi
done
