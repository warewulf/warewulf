package set

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/app/wwctl/flags"
	"github.com/warewulf/warewulf/internal/pkg/node"
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
		RunE:              CobraRunE(&vars),
		ValidArgsFunction: completions.Profiles(0), // no limit
	}
	vars.profileConf.CreateFlags(baseCmd)
	vars.profileDel.CreateDelFlags(baseCmd)
	vars.profileAdd.CreateAddFlags(baseCmd)
	flags.AddContainer(baseCmd, &(vars.profileConf.ImageName))
	baseCmd.PersistentFlags().BoolVarP(&vars.setYes, "yes", "y", false, "Set 'yes' to all questions asked")
	// register the command line completions
	if err := baseCmd.RegisterFlagCompletionFunc("image", completions.Images(0)); err != nil { // no limit
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("kernelversion", completions.ProfileKernelVersion); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("runtime", completions.Overlays); err != nil {
		panic(err)
	}
	if err := baseCmd.RegisterFlagCompletionFunc("wwinit", completions.Overlays); err != nil {
		panic(err)
	}
	return baseCmd
}
