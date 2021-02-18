package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open nodeDB: %s\n", err)
		os.Exit(1)
	}

	nodes, _ := nodeDB.FindAllNodes()

ARG_LOOP:
	for _, arg := range args {
		for _, n := range nodes {
			if n.ContainerName.Get() == arg {
				wwlog.Printf(wwlog.ERROR, "Container is configured for nodes, skipping: %s\n", arg)
				continue ARG_LOOP
			}
		}

		if container.ValidSource(arg) == false {
			wwlog.Printf(wwlog.ERROR, "Container name is not a valid source: %s\n", arg)
			continue
		}
		err := container.DeleteSource(arg)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not remove source: %s\n", arg)
		} else {
			fmt.Printf("Container has been deleted: %s\n", arg)
		}
	}

	return nil
}
