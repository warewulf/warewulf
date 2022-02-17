package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/kernel"
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
			if n.KernelVersion.Get() == arg {
				wwlog.Printf(wwlog.ERROR, "Kernel is configured for nodes, skipping: %s\n", arg)
				continue ARG_LOOP
			}
		}

		err := kernel.DeleteKernel(arg)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not delete kernel: %s\n", arg)
		} else {
			fmt.Printf("Kernel has been deleted: %s\n", arg)
		}
	}

	return nil
}
