package set

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/app/wwctl/flags"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	setNodeAll bool
	setYes     bool
	setForce   bool
	nodeConf   node.Node
	nodeDel    node.NodeConfDel
	nodeAdd    node.NodeConfAdd
}

func GetCommand() *cobra.Command {
	vars := variables{}
	vars.nodeConf = node.NewNode("")
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "set [OPTIONS] PATTERN",
		Short:                 "Configure node properties",
		Long:                  "This command sets configuration properties for nodes matching PATTERN.\n\nNote: use the string 'UNSET' to remove a configuration\n" + hostlist.Docstring,
		Aliases:               []string{"modify"},
		Args:                  cobra.MinimumNArgs(1), // require pattern as a mandatory arg
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction:     completions.Nodes,
	}

	vars.nodeConf.CreateFlags(baseCmd)
	vars.nodeAdd.CreateAddFlags(baseCmd)
	vars.nodeDel.CreateDelFlags(baseCmd)
	flags.AddContainer(baseCmd, &(vars.nodeConf.Profile.ImageName))
	flags.AddWwinit(baseCmd, &(vars.nodeConf.SystemOverlay))
	flags.AddRuntime(baseCmd, &(vars.nodeConf.RuntimeOverlay))
	baseCmd.PersistentFlags().BoolVarP(&vars.setNodeAll, "all", "a", false, "Set all nodes")
	baseCmd.PersistentFlags().BoolVarP(&vars.setYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&vars.setForce, "force", "f", false, "Force configuration (even on error)")
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("image", completions.Images); err != nil {
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
	if err := baseCmd.RegisterFlagCompletionFunc("profile", completions.Profiles); err != nil {
		panic(err)
	}

	return baseCmd
}
