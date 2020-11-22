package add

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed opening node database: %s\n", err)
		os.Exit(1)
	}

	nodeDB.AddGroup(args[0])


	nodeDB.Persist()

	return nil
}

