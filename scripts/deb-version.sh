#!/bin/sh
# Translate the upstream version string into a Debian-compatible upstream
# version. The transformation is the same as scripts/rpm-version.sh: a
# leading "rcN" suffix becomes "~rcN" so Debian sorts it as a pre-release,
# and "-" inside the git-describe suffix becomes ".".

if [ $# != 1 ]; then
    echo "Usage: deb-version.sh versionstring" >&2
    exit 2
fi

scripts/rpm-version.sh "$1"
