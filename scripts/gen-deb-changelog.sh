#!/bin/sh
# Generate a single-stanza debian/changelog for the given upstream version.
# The committed debian/changelog is a placeholder; `make deb` regenerates
# the real one each build because the version embeds a git-describe suffix.

if [ $# != 1 ]; then
    echo "Usage: gen-deb-changelog.sh versionstring" >&2
    exit 2
fi

VERSION=$1

cat <<EOF
warewulf (${VERSION}-1) unstable; urgency=medium

  * Build of warewulf ${VERSION}.

 -- Jonathon Anderson <janderson@ciq.com>  $(date -uR)
EOF
