#!/bin/bash

[ -z "${wwinit_root_device}" ] && return 0

get_stage() {
    stage="${1}"
    uri="${2:-${wwinit_uri}}"
    cacert="${3}"
    info "warewulf: loading stage: ${stage}"
    # Load runtime overlay from a static privledged port.
    # Others use default settings.
    localport=""
    if [ "${stage}" = "runtime" ]; then
        localport="--local-port 1-1023"
    fi
    cacert_opt=""
    if [ -n "${cacert}" ]; then
        cacert_opt="--cacert ${cacert}"
    fi
    (
        curl --location --silent --get ${localport} ${cacert_opt} \
            --retry 60 --retry-connrefused --retry-delay 1 \
            --data-urlencode "assetkey=${wwinit_assetkey}" \
            --data-urlencode "uuid=${wwinit_uuid}" \
            --data-urlencode "stage=${stage}" \
            --data-urlencode "compress=gz" \
            "${uri}" \
        | gzip -d \
        | cpio -ium --directory="${NEWROOT}"
    )
}

mkdir /tmp/wwinit
(
    # fetch the system overlay into /tmp/wwinit
    local NEWROOT=/tmp/wwinit
    get_stage "system" || die "Unable to load stage: system"
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

for stage in "image" "system"; do
    get_stage "${stage}" || die "Unable to load stage: ${stage}"
done

# Fetch runtime overlay (non-fatal)
# Source config from system overlay for TLS settings
. /tmp/wwinit/warewulf/config
cert_file="/tmp/wwinit/warewulf/tls/warewulf.crt"
if [ "${WWTLS}" = "true" ] && [ -f "$cert_file" ]; then
    # TLS enabled: build HTTPS URI using wwid from kernel cmdline
    # (mirrors wwclient URL construction in internal/app/wwclient/root.go)
    wwid=$(getarg wwid)
    runtime_uri="https://${WWIPADDR}:${WWTLSPORT}/provision/${wwid}"
    get_stage "runtime" "${runtime_uri}" "${cert_file}" || warn "warewulf: unable to load runtime overlay over HTTPS (ignored)"
else
    # No TLS: fetch runtime over HTTP
    get_stage "runtime" || warn "warewulf: unable to load runtime overlay (ignored)"
fi

# Copy /warewulf/run from initramfs to NEWROOT
# This preserves state files created by wwinit.d scripts (e.g., ignition marker)
if [ -d /tmp/wwinit/warewulf/run ]; then
    info "warewulf: preserving /warewulf/run to mounted root"
    mkdir -p "${NEWROOT}/warewulf"
    cp -a /tmp/wwinit/warewulf/run "${NEWROOT}/warewulf/"
fi
