package flags

import (
	"github.com/spf13/cobra"
)

func AddContainer(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "container", "C", "", "Set image name (backwards-compatibility)")
	cmd.Flags().Lookup("container").Hidden = true
	if err := cmd.Flags().MarkDeprecated("container", "use --image instead"); err != nil {
		panic(err)
	}
}

func AddWwinit(cmd *cobra.Command, dest *[]string) {
	cmd.Flags().StringSliceVar(dest, "wwinit", []string{}, "Set the system overlay")
	cmd.Flags().Lookup("wwinit").Hidden = true
	if err := cmd.Flags().MarkDeprecated("wwinit", "use --system-overlays instead"); err != nil {
		panic(err)
	}
}

func AddRuntime(cmd *cobra.Command, dest *[]string) {
	cmd.Flags().StringSliceVar(dest, "runtime", []string{}, "Set the runtime overlay")
	cmd.Flags().Lookup("runtime").Hidden = true
	if err := cmd.Flags().MarkDeprecated("runtime", "use --runtime-overlays instead"); err != nil {
		panic(err)
	}
}
