package upgrade

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/pkg/node"
	libupgrade "github.com/warewulf/warewulf/internal/pkg/upgrade"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

var (
	Command = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "upgrade [OPTIONS]",
		Short:                 "Upgrade an existing nodes.conf",
		Long: `Upgrades nodes.conf from a previous version of Warewulf 4 to a format
supported by the current version.`,
		RunE: UpgradeNodesConf,
	}

	addDefaults     bool
	replaceOverlays bool
	inputPath       string
	outputPath      string
)

func init() {
	Command.Flags().BoolVar(&addDefaults, "add-defaults", false, "Configure a default profile and set default node values")
	Command.Flags().BoolVar(&replaceOverlays, "replace-overlays", false, "Replace 'wwinit' and 'generic' overlays with their split replacements")
	Command.Flags().StringVarP(&inputPath, "input-path", "i", node.ConfigFile, "Path to a legacy nodes.conf")
	Command.Flags().StringVarP(&outputPath, "output-path", "o", node.ConfigFile, "Path to write the upgraded nodes.conf to")
}

func UpgradeNodesConf(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	legacy, err := libupgrade.Parse(data)
	if err != nil {
		return err
	}
	upgraded := legacy.Upgrade(addDefaults, replaceOverlays)
	if err := util.CopyFile(outputPath, outputPath+"-old"); err != nil {
		return err
	}
	return upgraded.PersistToFile(outputPath)
}
