package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/upgrade"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

var (
	inputPath  string
	outputPath string
)

func GetCommand() *cobra.Command {
	command := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "config [OPTIONS]",
		Short:                 "Upgrade an existing warewulf.conf",
		Long: `Upgrades warewulf.conf from a previous version of Warewulf 4 to a format
supported by the current version.`,
		RunE:              UpgradeNodesConf,
		Args:              cobra.NoArgs,
		ValidArgsFunction: completions.None,
	}
	command.Flags().StringVarP(&inputPath, "input-path", "i", "", "Path to a legacy warewulf.conf")
	command.Flags().StringVarP(&outputPath, "output-path", "o", "", "Path to write the upgraded warewulf.conf to")
	return command
}

func UpgradeNodesConf(cmd *cobra.Command, args []string) error {
	if inputPath == "" {
		inputPath = config.ConfigFile
	}
	if outputPath == "" {
		outputPath = config.ConfigFile
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	legacy, err := upgrade.ParseConfig(data)
	if err != nil {
		return err
	}
	upgraded := legacy.Upgrade()
	if outputPath == "-" {
		upgradedYaml, err := upgraded.Dump()
		if err != nil {
			return err
		}
		fmt.Print(string(upgradedYaml))
		return nil
	} else {
		if util.IsFile(outputPath) {
			if err := util.CopyFile(outputPath, outputPath+"-old"); err != nil {
				return err
			}
		}
		return upgraded.PersistToFile(outputPath)
	}
}
