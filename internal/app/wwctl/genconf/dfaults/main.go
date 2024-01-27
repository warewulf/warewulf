package dfaults

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	fmt.Println(node.FallBackConf)
	return
}
