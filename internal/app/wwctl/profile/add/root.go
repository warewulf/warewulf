package add

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/app/wwctl/flags"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	profileConf node.Profile
	profileAdd  node.NodeConfAdd
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	vars := variables{}
	vars.profileConf = node.NewProfile("")
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "add PROFILE",
		Short:                 "Add a new node profile",
		Long:                  "This command adds a new named PROFILE.",
		Aliases:               []string{"new", "create"},
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction:     completions.None,
	}
	vars.profileConf.CreateFlags(baseCmd)
	vars.profileAdd.CreateAddFlags(baseCmd)
	flags.AddContainer(baseCmd, &(vars.profileConf.ImageName))
	flags.AddWwinit(baseCmd, &(vars.profileConf.SystemOverlay))
	flags.AddRuntime(baseCmd, &(vars.profileConf.RuntimeOverlay))
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("image", completions.Images); err != nil { // no limit
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.ProfileKernelVersion); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("runtime-overlays", completions.OverlayList); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("runtime", completions.OverlayList); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("system-overlays", completions.OverlayList); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("wwinit", completions.OverlayList); err != nil {
		panic(err)
	}
	return baseCmd
}
