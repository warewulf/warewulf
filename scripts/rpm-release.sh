#!/bin/bash
# translate given version to rpm equivalent

if [ $# != 1 ]; then
    echo "Usage: rpm-release versionstring" >&2
    exit 2
fi

# Extract release from a colon-separated suffix:
#
# 3.4.2-rc.1:2  ->  2
release=$(echo "$1" | cut -sd: -f 2)
if [ "${release}" == "" ]
then
    echo 1
else
    echo "${release}"
fi
