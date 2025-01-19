package flags

import (
	"github.com/spf13/cobra"
)

func AddContainer(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "container", "C", "", "Set image name (backwards-compatibility)")
	cmd.Flags().Lookup("container").Hidden = true
}
