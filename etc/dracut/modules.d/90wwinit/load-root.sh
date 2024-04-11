#!/bin/bash

image=$(getarg wwinit.image)
system_overlay=$(getarg wwinit.system)
runtime_overlay=$(getarg wwinit.runtime)

info "Mounting tmpfs at $NEWROOT"
mount -t tmpfs tmpfs "$NEWROOT"

for archive in "${image}" "${system_overlay}" "${runtime_overlay}"
do
    if [ -n "${archive}" ]
    then
        info "Loading ${archive}"
        curl -L "${archive}" | gzip -d | cpio -im --directory="${NEWROOT}"
    fi
done
