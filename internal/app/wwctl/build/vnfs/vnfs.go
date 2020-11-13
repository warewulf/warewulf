package vnfs

import (
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

var buildForce bool

func Build(nodeList []assets.NodeInfo, force bool) error {
	set := make(map[string]int)

	wwlog.Printf(wwlog.INFO, "Building and Importing VNFS Images:\n")
	wwlog.SetIndent(4)

	buildForce = force

	for _, node := range nodeList {
		if node.Vnfs != "" {
			set[node.Vnfs]++
			wwlog.Printf(wwlog.DEBUG, "Node '%s' has VNFS '%s'\n", node.Fqdn, node.Vnfs)
		}
	}

	for uri := range set {
		v := vnfs.New(uri)
		wwlog.Printf(wwlog.VERBOSE, "VNFS found: %s (nodes: %d)\n", uri, set[uri])
		if strings.HasPrefix(uri, "/") {
			if strings.HasSuffix(uri, "tar.gz") {
				//wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball: %s\n", uri)
				wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball is not supported yet: %s\n", uri)

			} else {
				BuildContainerdir(v)
			}
		} else {
			BuildDocker(v)
		}
	}

	wwlog.SetIndent(0)

	return nil
}
