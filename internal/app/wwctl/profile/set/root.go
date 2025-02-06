package set

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/app/wwctl/flags"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

type variables struct {
	setNetDevDel string
	setDiskDel   string
	setPartDel   string
	setFsDel     string
	setNodeAll   bool
	setYes       bool
	setForce     bool
	partName     string
	diskName     string
	fsName       string
	profileConf  node.Profile
	profileDel   node.NodeConfDel
	profileAdd   node.NodeConfAdd
}

func GetCommand() *cobra.Command {
	vars := variables{}
	vars.profileConf = node.NewProfile("")

	baseCmd := &cobra.Command{
		Use:   "set [OPTIONS] [PROFILE ...]",
		Short: "Configure node profile properties",
		Long: "This command sets configuration properties for the node PROFILE(s).\n\n" +
			"Note: use the string 'UNSET' to remove a configuration",
		Aliases:           []string{"modify"},
		Args:              cobra.MinimumNArgs(0),
		RunE:              CobraRunE(&vars),
		ValidArgsFunction: completions.Profiles,
	}
	vars.profileConf.CreateFlags(baseCmd)
	vars.profileDel.CreateDelFlags(baseCmd)
	vars.profileAdd.CreateAddFlags(baseCmd)
	flags.AddContainer(baseCmd, &(vars.profileConf.ImageName))
	baseCmd.PersistentFlags().BoolVarP(&vars.setYes, "yes", "y", false, "Set 'yes' to all questions asked")
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("image", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		list, _ := image.ListSources()
		return list, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		log.Println(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.ProfileKernelVersion); err != nil {
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
	return baseCmd
}
