package set

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	var c int

	n, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if SetVnfs != "" {
		fmt.Printf("Setting vnfs to: %s\n", SetVnfs)
		n = n.SetNodeVal("n0000", "vnfs", SetVnfs)
	}

	fmt.Printf("set count: %d\n", c)

	a, err := n.FindByHwaddr("00:0c:29:23:8b:48")

	fmt.Printf("VNFS: %s\n", a.Vnfs)

	//n.Persist()

	return nil
}