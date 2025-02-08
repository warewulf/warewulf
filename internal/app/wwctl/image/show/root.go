package show

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "show [OPTIONS] IMAGE",
		Short:                 "Show root fs dir for image",
		Long: `Shows the base directory for the chroot of the given image.
More information about the image can be shown with the '-a' option.`,
		RunE:              CobraRunE,
		ValidArgsFunction: completions.Images(0), // no limit
		Args:              cobra.ExactArgs(1),
	}
	ShowAll bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&ShowAll, "all", "a", false, "Show all information about an image")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
