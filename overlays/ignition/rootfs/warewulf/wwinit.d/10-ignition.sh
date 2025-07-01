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
    ignition --config-cache="${PREFIX}/warewulf/ignition.json" --platform=metal --stage=disks || die "warewulf: ignition: failed to partition/format disk"
else
    info "warewulf: ignition not found"
fi
