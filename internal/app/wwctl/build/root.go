package build

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		Use:                "build",
		Short:              "Warewulf build subcommand",
		Long:               "Warewulf build is used to build VNFS, kernel, and system-overlay objects for\n" +
							"provisioning. The default usage will be to build and/or update anything\n" +
							"that seems to be needed.",
		RunE:				CobraRunE,
	}

	buildVnfs bool
	buildKernel bool
	buildRuntimeOverlay bool
	buildSystemOverlay bool
	buildAll bool
	buildForce bool
)

func init() {
	baseCmd.PersistentFlags().BoolVarP(&buildVnfs, "vnfs", "V", false, "Build and/or update VNFS images.")
	baseCmd.PersistentFlags().BoolVarP(&buildKernel, "kernel", "K", false, "Build and/or update Kernel images.")
	baseCmd.PersistentFlags().BoolVarP(&buildRuntimeOverlay, "runtime", "R", false, "Build and/or update runtime overlays")
	baseCmd.PersistentFlags().BoolVarP(&buildSystemOverlay, "system", "S", false, "Build and/or update system overlays")
	baseCmd.PersistentFlags().BoolVarP(&buildAll, "all", "A", false, "Build and/or update all components")
	baseCmd.PersistentFlags().BoolVarP(&buildForce, "force", "f", false, "Force build even if nothing has been updated.")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
