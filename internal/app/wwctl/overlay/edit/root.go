package edit

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "edit",
		Short:              "Edit Warewulf Overlay files",
		Long:               "Warewulf edit overlay files",
		RunE:				CobraRunE,
	}
	SystemOverlay bool
	ListFiles bool

)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&SystemOverlay, "system", "s", false, "Show system overlays instead of runtime")
	baseCmd.PersistentFlags().BoolVarP(&ListFiles, "files", "f", false, "List files contained within a given overlay")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}


func CobraRunE(cmd *cobra.Command, args []string) error {
	fmt.Printf("Edit: Hello World\n")
	return nil
}