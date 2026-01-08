package unset

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	unsetYes    bool
	unsetForce  bool
	unsetFields map[string]*bool
	profileConf node.Profile
	profileAdd  node.NodeConfAdd
}

func GetCommand() *cobra.Command {
	vars := variables{}
	vars.profileConf = node.NewProfile("")

	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "unset [OPTIONS] PROFILE...",
		Short:                 "Unset/clear profile properties",
		Long:                  "Unsets configuration properties for the specified PROFILE(s).",
		Args:                  cobra.MinimumNArgs(1),
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction:     completions.Profiles,
	}

	// Create unset flags and store map
	vars.unsetFields = vars.profileConf.CreateUnsetFlags(baseCmd)

	// Add --netname for specifying which network device
	vars.profileAdd.CreateAddFlags(baseCmd)

	// Add control flags
	baseCmd.PersistentFlags().BoolVarP(&vars.unsetYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&vars.unsetForce, "force", "f", false, "Force configuration (even on error)")

	// Register completions for flags created by CreateUnsetFlags()
	if err := baseCmd.RegisterFlagCompletionFunc("image", completions.Images); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.ProfileKernelVersion); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("runtime-overlays", completions.OverlayList); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("system-overlays", completions.OverlayList); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("profile", completions.Profiles); err != nil {
		panic(err)
	}

	return baseCmd
}
