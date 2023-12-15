package delete

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open nodeDB: %s", err)
		os.Exit(1)
	}

	nodes, _ := nodeDB.FindAllNodes()

ARG_LOOP:
	for _, arg := range args {
		for _, n := range nodes {
			if n.Kernel.Override == arg {
				wwlog.Error("Kernel is configured for nodes, skipping: %s", arg)
				continue ARG_LOOP
			}
		}

		err := kernel.DeleteKernel(arg)
		if err != nil {
			wwlog.Error("Could not delete kernel: %s", arg)
		} else {
			fmt.Printf("Kernel has been deleted: %s\n", arg)
		}
	}

	return nil
}
