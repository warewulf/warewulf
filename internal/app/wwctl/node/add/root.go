package add

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/app/wwctl/flags"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
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
	}
	vars.nodeConf.CreateFlags(baseCmd)
	vars.nodeAdd.CreateAddFlags(baseCmd)
	flags.AddContainer(baseCmd, &(vars.nodeConf.Profile.ImageName))
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("image", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := image.ListSources()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.NodeKernelVersion); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("runtime", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list := overlay.FindOverlays()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("wwinit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list := overlay.FindOverlays()
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

	// GetRootCommand returns the root cobra.Command for the application.
	return baseCmd
}
