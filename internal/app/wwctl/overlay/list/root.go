package list

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:                "list",
		Short:              "List Warewulf Overlays",
		Long:               "Warewulf List overlay",
		RunE:				CobraRunE,
	}

)

func init() {

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return listCmd
}


func CobraRunE(cmd *cobra.Command, args []string) error {

	fmt.Printf("List: Hello World\n")
	return nil
}