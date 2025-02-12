package add

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/app/wwctl/flags"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

// Holds the variables which are needed in CobraRunE
type variables struct {
	nodeConf node.Node
	nodeAdd  node.NodeConfAdd
}

// Returns the newly created command
func GetCommand() *cobra.Command {
	vars := variables{}
	vars.nodeConf = node.NewNode("")
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "add [OPTIONS] NODENAME",
		Short:                 "Add new node to Warewulf",
		Long:                  "This command will add a new node named NODENAME to Warewulf.",
		Aliases:               []string{"new", "create"},
		RunE:                  CobraRunE(&vars),
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     cobra.FixedCompletions(nil, cobra.ShellCompDirectiveNoFileComp),
	}
	vars.nodeConf.CreateFlags(baseCmd)
	vars.nodeAdd.CreateAddFlags(baseCmd)
	flags.AddContainer(baseCmd, &(vars.nodeConf.Profile.ImageName))
	flags.AddWwinit(baseCmd, &(vars.nodeConf.SystemOverlay))
	flags.AddRuntime(baseCmd, &(vars.nodeConf.RuntimeOverlay))
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("image", completions.Images); err != nil { // no limit
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.NodeKernelVersion); err != nil {
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
	if err := baseCmd.RegisterFlagCompletionFunc("profile", completions.Profiles); err != nil { // no limit
		panic(err)
	}

	// GetRootCommand returns the root cobra.Command for the application.
	return baseCmd
}
