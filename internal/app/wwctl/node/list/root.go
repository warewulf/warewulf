package list

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

type variables struct {
	showNet  bool
	showIpmi bool
	showAll  bool
	showLong bool
	showYaml bool
	showJson bool
	format   string
}

func GetCommand() *cobra.Command {
	vars := variables{}
	baseCmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "list [OPTIONS] [PATTERN]",
		Short:                 "List nodes",
		Long: "This command lists all configured nodes. Optionally, it will list only\n" +
			"nodes matching a PATTERN.\n" + hostlist.Docstring,
		RunE:              CobraRunE(&vars),
		Aliases:           []string{"ls"},
		ValidArgsFunction: completions.Nodes,
		Args:              cobra.ArbitraryArgs,
	}
	baseCmd.PersistentFlags().BoolVarP(&vars.showNet, "net", "n", false, "Show node network configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showIpmi, "ipmi", "i", false, "Show node IPMI configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showAll, "all", "a", false, "Show all node configurations")
	baseCmd.PersistentFlags().BoolVarP(&vars.showLong, "long", "l", false, "Show long or wide format")
	baseCmd.PersistentFlags().BoolVarP(&vars.showYaml, "yaml", "y", false, "Show yaml format")
	baseCmd.PersistentFlags().BoolVarP(&vars.showJson, "json", "j", false, "Show json format")
	baseCmd.PersistentFlags().StringVarP(&vars.format, "format", "f", "", "Show formatted output, as the format must be a template")

	return baseCmd
}
