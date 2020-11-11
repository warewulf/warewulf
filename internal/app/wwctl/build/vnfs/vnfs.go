package vnfs

import (
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"strings"
)

var buildForce bool

func Build(nodeList []assets.NodeInfo, force bool) (error) {
	set := make(map[string]int)

	wwlog.Printf(wwlog.INFO, "Importing VNFS Images:\n")
	wwlog.SetIndent(4)

	buildForce = force

	for _, node := range nodeList {
		if node.Vnfs != "" {
			set[node.Vnfs] ++
			wwlog.Printf(wwlog.DEBUG, "Node '%s' has VNFS '%s'\n", node.Fqdn, node.Vnfs)
		}
	}

	for uri := range set {
		v := vnfs.New(uri)
		wwlog.Printf(wwlog.VERBOSE, "VNFS found: %s (nodes: %d)\n", uri, set[uri])
		if strings.HasPrefix(uri, "docker://") {
			BuildDocker(v)

		} else if strings.HasPrefix(uri, "docker-daemon://") {
			//wwlog.Printf(wwlog.INFO, "Building VNFS from Docker service: %s\n", uri)
			wwlog.Printf(wwlog.INFO, "Building VNFS from Docker service is not supported yet: %s\n", uri)

		} else if strings.HasPrefix(uri, "/") {
			if strings.HasSuffix(uri, "tar.gz") {
				//wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball: %s\n", uri)
				wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball is not supported yet: %s\n", uri)

			} else {
				BuildContainerdir(v)
			}
		}
	}

	wwlog.SetIndent(0)

	return nil
}