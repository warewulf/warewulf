package dfaults

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	fmt.Println(node.FallBackConf)
	return
}
