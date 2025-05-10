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

if [ "${wwinit_root_device}" != "tmpfs" ]; then
    mkdir /tmp/wwinit
    NEWROOT=/tmp/wwinit get_stage "system"
    if [ -x /usr/bin/ignition ]; then
        /usr/bin/ignition --root "${NEWROOT}" --config-cache=/tmp/wwinit/warewulf/ignition.json --platform=metal --stage=disks || die "warewulf: failed to partition/format disk"
    else
        info "warewulf: /usr/bin/ignition not found. Assuming ${wwinit_root_device} already prepared."
    fi
fi

info "warewulf: Mounting ${wwinit_root_device} at ${NEWROOT}"
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
