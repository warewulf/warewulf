#!/bin/sh

set -e
set -u

if [ -d "vendor" ]; then
    echo "Please remove vendor directory before running this script"
    exit 255
fi

if [ ! -f "go.mod" ]; then
    echo "This script must be called from the project root directory,"
    echo "i.e. as scripts/update-license-dependencise.sh"
    exit 255
fi

# Exclude ourselves
exclude="github.com/warewulf/warewulf"

# go-licenses only looks for a license file in the root of a
# dependency's own Go module. Dependencies that declare their license
# some other way are recorded here, as
# "<license URL>,<license name>" keyed by package.
override_license() {
    case "$1" in
	github.com/rootless-containers/proto/go-proto)
	    # A nested module of github.com/rootless-containers/proto,
	    # which is Apache-2.0 per its README and root COPYING.
	    echo "https://github.com/rootless-containers/proto/blob/master/README.md#license,Apache-2.0"
	    ;;
    esac
}

# Ensure a constant sort order
export LC_ALL=C

${GOLANG_LICENSES:-go-licenses} csv ./... \
    | grep -v -E "${exclude}" \
    | while IFS="," read -r dep url license; do
	override=$(override_license "${dep}")
	if [ -n "${override}" ]; then
	    echo "${dep},${override}"
	else
	    echo "${dep},${url},${license}"
	fi
    done \
    | sort -k3,3 -k1,1 -t, > LICENSE_DEPENDENCIES.csv

# Warn, rather than fail, so that a new dependency can still be
# recorded while its license is being determined.
if grep -q ",Unknown$" LICENSE_DEPENDENCIES.csv; then
    echo "Warning: could not determine the license of these dependencies." >&2
    echo "Identify them manually and add them to override_license in $0:" >&2
    grep ",Unknown$" LICENSE_DEPENDENCIES.csv | cut -d, -f1 | sed 's/^/  /' >&2
fi

# Header for the markdown file
cat <<-'EOF' >LICENSE_DEPENDENCIES.md
# Dependency Licenses

This project uses a number of dependencies, in accordance with their
own license terms. These dependencies are managed via the project
`go.mod` and `go.sum` files, and included in a `vendor/` directory in
our official source tarballs.

A full build or package of Warewulf uses all dependencies listed
below. If you `import "github.com/warewulf/warewulf"` into your own
project then you may use a subset of them.

The dependencies and their licenses are as follows:

EOF

while IFS="," read -r dep url license; do
    {
	echo "## ${dep}"
	echo ""
	echo "**License:** ${license}"
	echo ""
    } >>LICENSE_DEPENDENCIES.md

    # go-licenses can't work out the web url for non-github projects.
    # Fall back to using the dependency URL as a project URL
    if [ "${url}" = "Unknown" ]; then
	echo "**Project URL:** <https://${dep}>" >>LICENSE_DEPENDENCIES.md
    else
	echo "**License URL:** <${url}>" >>LICENSE_DEPENDENCIES.md
    fi
    echo "" >>LICENSE_DEPENDENCIES.md
done <LICENSE_DEPENDENCIES.csv

# Clean up
rm LICENSE_DEPENDENCIES.csv
