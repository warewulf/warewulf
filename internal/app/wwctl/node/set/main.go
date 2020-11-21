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
	var count int
	var nodes []node.NodeInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		nodes, err = nodeDB.SearchByNameList(args)
	} else {
		cmd.Usage()
		os.Exit(1)
	}
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}
	for _, n := range nodes {
		if SetVnfs != "" {
			fmt.Printf("Setting vnfs to: %s\n", SetVnfs)
			c, _ := nodeDB.SetNodeVal(n.Id, "vnfs", SetVnfs)
			count += c
		}
		if SetKernel != "" {
			fmt.Printf("Setting kernel to: %s\n", SetVnfs)
			c, _ := nodeDB.SetNodeVal(n.Id, "kernel", SetKernel)
			count += c
		}
	}

	fmt.Printf("set count: %d\n", count)

//	nodeDB.AddGroup("moo")
	nodeDB.AddNode("moo", "node01")

	nodeDB.Persist()

	return nil
}