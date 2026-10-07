package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/upgrade"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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
	// Warewulf no longer applies a `host` overlay implicitly. If a site
	// `host` overlay is still present, retain it explicitly so that the
	// site's existing customizations keep being applied after the upgrade.
	// A stale distribution copy (e.g. left by `make install`) is ignored.
	retainLegacyHostOverlay := false
	if o, err := overlay.Get("host"); err == nil && o.IsSiteOverlay() {
		retainLegacyHostOverlay = true
		wwlog.Warn("Found a legacy `host` overlay: adding it to each service's `overlays`" +
			" so that it continues to be applied. Migrate its files into the per-service" +
			" host overlays and remove those entries.")
	} else if err != nil && !errors.Is(err, overlay.ErrDoesNotExist) {
		return err
	}
	upgraded := legacy.Upgrade(retainLegacyHostOverlay)
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
