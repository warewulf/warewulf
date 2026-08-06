#!/bin/sh

if ! command -v info >/dev/null; then
    info() {
        printf '%s\n' "$*"
        return 0
    }
fi

if ! command -v warn >/dev/null; then
    warn() {
        printf '%s\n' "$*" >&2
        return 0
    }
fi

scriptdir="${PREFIX}/warewulf/wwinit.d"
if [ -d "${scriptdir}" ]; then
    info "warewulf: running scripts in ${scriptdir}..."
    for script in "${scriptdir}"/*; do
        [ -f "${script}" ] || continue
        name="${script##*/}"
        info "warewulf: ${name}"
        PREFIX=$PREFIX sh "${script}"
        status=$?
        if [ "${status}" -ne 0 ]; then
            warn "warewulf: ${name} exited with status ${status}"
            exit "${status}"
        fi
    done
else
    info "warewulf: ${scriptdir} not found"
fi
