package man

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	header := &doc.GenManHeader{
		Title:   "WWCTL",
		Section: "1",
	}
	err = doc.GenManTree(cmd.Parent().Parent(), header, args[0])
	return
}
