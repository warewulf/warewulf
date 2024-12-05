package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/pkg/config"
	libupgrade "github.com/warewulf/warewulf/internal/pkg/upgrade"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

var (
	Command = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "config [OPTIONS]",
		Short:                 "Upgrade an existing warewulf.conf",
		Long: `Upgrades warewulf.conf from a previous version of Warewulf 4 to a format
supported by the current version.`,
		RunE: UpgradeNodesConf,
	}

	inputPath  string
	outputPath string
)

func init() {
	Command.Flags().StringVarP(&inputPath, "input-path", "i", config.ConfigFile, "Path to a legacy warewulf.conf")
	Command.Flags().StringVarP(&outputPath, "output-path", "o", config.ConfigFile, "Path to write the upgraded warewulf.conf to")
}

func UpgradeNodesConf(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	legacy, err := libupgrade.ParseConfig(data)
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
		if err := util.CopyFile(outputPath, outputPath+"-old"); err != nil {
			return err
		}
		return upgraded.PersistToFile(outputPath)
	}
}
