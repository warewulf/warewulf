#!/bin/sh

if ! command -v info >/dev/null; then
    info() {
        printf '%s\n' "$*"
    }
fi

scriptdir="${PREFIX}/warewulf/wwinit.d"
if [ -d "${scriptdir}" ]; then
    info "warewulf: running scripts in ${scriptdir}..."
    ls -1 "${scriptdir}/" | while read -r name; do
        info "warewulf: ${name}"
        PREFIX=$PREFIX sh "${scriptdir}/${name}"
    done
else
    info "warewulf: ${scriptdir} not found"
fi
