#!/bin/sh

PATH=$PATH:/sbin:/usr/sbin:/bin:/usr/bin

if ! command -v info >/dev/null; then
    info() {
        printf '%s\n' "$*"
    }
fi

if ! command -v die >/dev/null; then
    die() {
        printf '%s\n' "$*" >&2
        exit 1
    }
fi

if command -v ignition >/dev/null; then :
    info "warewulf: ignition: partition and format disks"
    if ignition --config-cache="${PREFIX}/warewulf/ignition.json" --platform=metal --stage=disks; then
        # Create marker file to signal successful completion
        # This prevents the systemd service from running ignition again after switch_root
        mkdir -p "${PREFIX}/warewulf/run"
        echo "ignition run by 10-ignition.sh" >"${PREFIX}/warewulf/run/.ignition-done"
    else
        die "warewulf: ignition: failed to partition/format disk"
    fi
else
    info "warewulf: ignition not found"
fi
