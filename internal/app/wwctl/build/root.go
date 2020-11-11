package build

import (
	"github.com/spf13/cobra"
)

var (
	buildCmd = &cobra.Command{
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
	buildCmd.PersistentFlags().BoolVarP(&buildVnfs, "vnfs", "V", false, "Build and/or update VNFS images.")
	buildCmd.PersistentFlags().BoolVarP(&buildKernel, "kernel", "K", false, "Build and/or update Kernel images.")
	buildCmd.PersistentFlags().BoolVarP(&buildRuntimeOverlay, "runtime", "R", false, "Build and/or update runtime overlays")
	buildCmd.PersistentFlags().BoolVarP(&buildSystemOverlay, "system", "S", false, "Build and/or update system overlays")
	buildCmd.PersistentFlags().BoolVarP(&buildAll, "all", "A", false, "Build and/or update all components")
	buildCmd.PersistentFlags().BoolVarP(&buildForce, "force", "f", false, "Force build even if nothing has been updated.")

}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return buildCmd
}
