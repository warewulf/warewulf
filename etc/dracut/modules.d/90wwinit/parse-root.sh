#!/bin/sh
# root=cpio+http://<server>/path/to/cpio

[ -z "$root" ] && root=$(getarg root=)

if [ "${root}" = "wwinit" ]; then
    rootok=1
fi
