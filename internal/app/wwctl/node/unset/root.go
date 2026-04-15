package unset

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	wwctlunset "github.com/warewulf/warewulf/internal/app/wwctl/unset"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func GetCommand() *cobra.Command {
	vars := wwctlunset.Vars{}
	nodeConf := node.NewNode("")

	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "unset [OPTIONS] PATTERN",
		Short:                 "Unset/clear node properties",
		Long:                  "Unsets configuration properties for nodes matching PATTERN.\n\n" + hostlist.Docstring,
		Args:                  cobra.MinimumNArgs(1),
		RunE:                  CobraRunE(&vars),
		ValidArgsFunction:     completions.Nodes,
	}

	// Create unset flags and store maps
	vars.UnsetFields, vars.UnsetScopes = nodeConf.CreateUnsetFlags(baseCmd)

	// Add scoping flags for specifying which sub-entity to modify
	baseCmd.PersistentFlags().StringVar(&vars.Netname, "netname", "default", "network which is modified")
	baseCmd.PersistentFlags().StringVar(&vars.Diskname, "diskname", "", "disk to modify")
	baseCmd.PersistentFlags().StringVar(&vars.Partname, "partname", "", "partition to modify (requires --diskname)")
	baseCmd.PersistentFlags().StringVar(&vars.Fsname, "fsname", "", "filesystem to modify")

	// Add tag deletion flags
	baseCmd.PersistentFlags().StringSliceVar(&vars.Tags, "tag", []string{}, "Unset tags")
	baseCmd.PersistentFlags().StringSliceVar(&vars.IpmiTags, "ipmitag", []string{}, "Unset IPMI tags")
	baseCmd.PersistentFlags().StringSliceVar(&vars.NetTags, "nettag", []string{}, "Unset network tags")

	// Add object deletion flags
	baseCmd.PersistentFlags().StringSliceVar(&vars.NetDel, "net", []string{}, "Unset network device by name")
	baseCmd.PersistentFlags().StringSliceVar(&vars.DiskDel, "disk", []string{}, "Unset disk by name")
	baseCmd.PersistentFlags().StringSliceVar(&vars.PartDel, "part", []string{}, "Unset partition by name")
	baseCmd.PersistentFlags().StringSliceVar(&vars.FsDel, "fs", []string{}, "Unset filesystem by name")

	// Add control flags
	baseCmd.PersistentFlags().BoolVarP(&vars.UnsetYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&vars.UnsetForce, "force", "f", false, "Force configuration (even on error)")

	return baseCmd
}
