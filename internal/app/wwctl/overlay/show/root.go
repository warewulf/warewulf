package show

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "show",
		Short:              "Show Warewulf Overlay objects",
		Long:               "Warewulf show overlay objects",
		RunE:				CobraRunE,
	}

)

func init() {

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}


func CobraRunE(cmd *cobra.Command, args []string) error {
	fmt.Printf("Show: Hello World\n")
	return nil
}