#!/bin/bash

info "Mounting tmpfs at $NEWROOT"
mount -t tmpfs -o mpol=interleave ${wwinit_tmpfs_size_option} tmpfs "$NEWROOT"

for archive in "${wwinit_container}" "${wwinit_system}" "${wwinit_runtime}"
do
    if [ -n "${archive}" ]
    then
        info "Loading ${archive}"
        # Load runtime overlay from a static privledged port.
        # Others use default settings.
        localport=""
        if [[ "${archive}" == "${wwinit_runtime}" ]]
        then
            localport="--local-port 1-1023"
        fi
        (curl --silent ${localport} -L "${archive}" | gzip -d | cpio -im --directory="${NEWROOT}") || die "Unable to load ${archive}"
    fi
done
