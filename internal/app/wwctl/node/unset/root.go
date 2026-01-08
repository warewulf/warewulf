package unset

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	unsetNodeAll bool
	unsetYes     bool
	unsetForce   bool
	unsetFields  map[string]*bool
	nodeConf     node.Node        // For traversing struct
	nodeAdd      node.NodeConfAdd // For --netname
}

func GetCommand() *cobra.Command {
	vars := variables{}
	vars.nodeConf = node.NewNode("")

	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "unset [OPTIONS] PATTERN",
		Short:                 "Unset/clear node properties",
		Long:                  "Unsets configuration properties for nodes matching PATTERN.\n\n" + hostlist.Docstring,
		Args:                  cobra.MinimumNArgs(1),
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction:     completions.Nodes,
	}

	// Create unset flags and store map
	vars.unsetFields = vars.nodeConf.CreateUnsetFlags(baseCmd)

	// Add --netname for specifying which network device
	vars.nodeAdd.CreateAddFlags(baseCmd)

	// Add control flags
	baseCmd.PersistentFlags().BoolVarP(&vars.unsetNodeAll, "all", "a", false, "Unset all nodes")
	baseCmd.PersistentFlags().BoolVarP(&vars.unsetYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&vars.unsetForce, "force", "f", false, "Force configuration (even on error)")

	// Register completions for flags created by CreateUnsetFlags()
	// Note: We don't use the deprecated aliases (--runtime, --wwinit, --container)
	// that the set command has, since unset is a new command
	if err := baseCmd.RegisterFlagCompletionFunc("image", completions.Images); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.NodeKernelVersion); err != nil {
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
