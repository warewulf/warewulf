package add

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		os.Exit(1)
	}

	for _, a := range args {
		err = nodeDB.AddNode(SetGroup, a)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added node: %s\n", a)
	}

	nodeDB.Persist()

	return nil
}