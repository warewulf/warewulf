package list

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:   "list [flags] <overlay kind> <overlay name>",
		Short: "List Warewulf Overlays and files",
		Long: "This command will show you information about Warewulf overlays and the\n" +
			"files contained within.",
		RunE:    CobraRunE,
		Aliases: []string{"ls"},
	}
	ListContents bool
	ListLong     bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ListContents, "all", "a", false, "List the contents of overlays")
	baseCmd.PersistentFlags().BoolVarP(&ListLong, "long", "l", false, "List 'long' of all overlay contents")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
