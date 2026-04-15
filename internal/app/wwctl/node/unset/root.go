package unset

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

type variables struct {
	unsetYes    bool
	unsetForce  bool
	unsetFields map[string]*bool
	unsetScopes map[string]string
	netname     string
	diskname    string
	partname    string
	fsname      string
	tags        []string
	ipmiTags    []string
	netTags     []string
	netDel      []string
	diskDel     []string
	partDel     []string
	fsDel       []string
}

func GetCommand() *cobra.Command {
	vars := variables{}
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
	vars.unsetFields, vars.unsetScopes = nodeConf.CreateUnsetFlags(baseCmd)

	// Add scoping flags for specifying which sub-entity to modify
	baseCmd.PersistentFlags().StringVar(&vars.netname, "netname", "default", "network which is modified")
	baseCmd.PersistentFlags().StringVar(&vars.diskname, "diskname", "", "disk to modify")
	baseCmd.PersistentFlags().StringVar(&vars.partname, "partname", "", "partition to modify (requires --diskname)")
	baseCmd.PersistentFlags().StringVar(&vars.fsname, "fsname", "", "filesystem to modify")

	// Add tag deletion flags
	baseCmd.PersistentFlags().StringSliceVar(&vars.tags, "tag", []string{}, "Unset tags")
	baseCmd.PersistentFlags().StringSliceVar(&vars.ipmiTags, "ipmitag", []string{}, "Unset IPMI tags")
	baseCmd.PersistentFlags().StringSliceVar(&vars.netTags, "nettag", []string{}, "Unset network tags")

	// Add object deletion flags
	baseCmd.PersistentFlags().StringSliceVar(&vars.netDel, "net", []string{}, "Unset network device by name")
	baseCmd.PersistentFlags().StringSliceVar(&vars.diskDel, "disk", []string{}, "Unset disk by name")
	baseCmd.PersistentFlags().StringSliceVar(&vars.partDel, "part", []string{}, "Unset partition by name")
	baseCmd.PersistentFlags().StringSliceVar(&vars.fsDel, "fs", []string{}, "Unset filesystem by name")

	// Add control flags
	baseCmd.PersistentFlags().BoolVarP(&vars.unsetYes, "yes", "y", false, "Set 'yes' to all questions asked")
	baseCmd.PersistentFlags().BoolVarP(&vars.unsetForce, "force", "f", false, "Force configuration (even on error)")

	return baseCmd
}
