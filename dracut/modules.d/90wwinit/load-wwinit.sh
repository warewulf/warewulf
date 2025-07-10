#!/bin/bash

get_stage() {
    stage="${1}"
    info "warewulf: loading stage: ${stage}"
    # Load runtime overlay from a static privledged port.
    # Others use default settings.
    localport=""
    if [ "${stage}" = "runtime" ]; then
        localport="--local-port 1-1023"
    fi
    (
        curl --location --silent --get ${localport} \
            --retry 60 --retry-connrefused --retry-delay 1 \
            --data-urlencode "assetkey=${wwinit_assetkey}" \
            --data-urlencode "uuid=${wwinit_uuid}" \
            --data-urlencode "stage=${stage}" \
            --data-urlencode "compress=gz" \
            "${wwinit_uri}" \
        | gzip -d \
        | cpio -ium --directory="${NEWROOT}"
    ) || die "Unable to load stage: ${stage}"
}

mkdir /tmp/wwinit
(
    # fetch the system overlay into /tmp/wwinit
    local NEWROOT=/tmp/wwinit
    get_stage "system"
)
if [ -x /tmp/wwinit/warewulf/run-wwinit.d ]; then
        PREFIX=/tmp/wwinit /tmp/wwinit/warewulf/run-wwinit.d
fi

info "warewulf: mounting ${wwinit_root_device} at ${NEWROOT}"
(
    if [ "${wwinit_root_device}" = "tmpfs" ]; then
        mount -t tmpfs -o mpol=interleave ${wwinit_tmpfs_size_option} "${wwinit_root_device}" "${NEWROOT}"
    else
        mount "${wwinit_root_device}" "${NEWROOT}"
    fi
) || die "warewulf: failed to mount ${wwinit_root_device} at ${NEWROOT}"

for stage in "image" "system" "runtime"; do
    get_stage "${stage}"
done
