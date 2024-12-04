package set

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
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
		Long:                  "This command sets configuration properties for nodes matching PATTERN.\n\nNote: use the string 'UNSET' to remove a configuration",
		Aliases:               []string{"modify"},
		Args:                  cobra.MinimumNArgs(1), // require pattern as a mandatory arg
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			nodeDB, _ := node.New()
			nodes := nodeDB.ListAllNodes()
			return nodes, cobra.ShellCompDirectiveNoFileComp
		},
	}

	vars.nodeConf.CreateFlags(baseCmd)
	vars.nodeAdd.CreateAddFlags(baseCmd)
	vars.nodeDel.CreateDelFlags(baseCmd)
	baseCmd.PersistentFlags().BoolVarP(&vars.setNodeAll, "all", "a", false, "Set all nodes")
	baseCmd.PersistentFlags().BoolVarP(&vars.setYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&vars.setForce, "force", "f", false, "Force configuration (even on error)")
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("container", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := container.ListSources()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.NodeKernelVersion); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("runtime", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("wwinit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var list []string
		nodeDB, _ := node.New()
		profiles, _ := nodeDB.FindAllProfiles()
		for _, profile := range profiles {
			list = append(list, profile.Id())
		}
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}

	return baseCmd
}
