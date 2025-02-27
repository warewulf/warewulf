#!/bin/bash

info "Mounting tmpfs at $NEWROOT"
mount -t tmpfs -o mpol=interleave ${wwinit_tmpfs_size_option} tmpfs "$NEWROOT"

for stage in "image" "system" "runtime"
do
    info "Loading stage: ${stage}"
    # Load runtime overlay from a static privledged port.
    # Others use default settings.
    localport=""
    if [[ "${stage}" == "runtime" ]]
    then
        localport="--local-port 1-1023"
    fi
    (
        curl --location --silent --get ${localport} \
            --retry 60 --retry-delay 1 --retry-all-errors \
            --data-urlencode "assetkey=${wwinit_assetkey}" \
            --data-urlencode "uuid=${wwinit_uuid}" \
            --data-urlencode "stage=${stage}" \
            --data-urlencode "compress=gz" \
            "${wwinit_uri}" \
        | gzip -d \
        | cpio -im --directory="${NEWROOT}"
    ) || die "Unable to load stage: ${stage}"
done
