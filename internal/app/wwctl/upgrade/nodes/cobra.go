package nodes

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	libupgrade "github.com/warewulf/warewulf/internal/pkg/upgrade"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

var (
	Command = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "nodes [OPTIONS]",
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
	controller := warewulfconf.Get()
	Command.Flags().BoolVar(&addDefaults, "add-defaults", false, "Configure a default profile and set default node values")
	Command.Flags().BoolVar(&replaceOverlays, "replace-overlays", false, "Replace 'wwinit' and 'generic' overlays with their split replacements")
	Command.Flags().StringVarP(&inputPath, "input-path", "i", path.Join(controller.Paths.Sysconfdir, "warewulf/nodes.conf"), "Path to a legacy nodes.conf")
	Command.Flags().StringVarP(&outputPath, "output-path", "o", path.Join(controller.Paths.Sysconfdir, "warewulf/nodes.conf"), "Path to write the upgraded nodes.conf to")
	if err := Command.MarkFlagRequired("add-defaults"); err != nil {
		panic(err)
	}
	if err := Command.MarkFlagRequired("replace-overlays"); err != nil {
		panic(err)
	}
}

func UpgradeNodesConf(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	legacy, err := libupgrade.ParseNodes(data)
	if err != nil {
		return err
	}
	upgraded := legacy.Upgrade(addDefaults, replaceOverlays)
	if outputPath == "-" {
		upgradedYaml, err := upgraded.Dump()
		if err != nil {
			return err
		}
		fmt.Print(string(upgradedYaml))
		return nil
	} else {
		if err := util.CopyFile(outputPath, outputPath+"-old"); err != nil {
			return err
		}
		return upgraded.PersistToFile(outputPath)
	}
}
