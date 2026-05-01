package capture

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "capture",
		Short:                 "Capture deterministic differences between source and baseline",
		Long:                  "Capture deterministic differences between source and baseline trees",
		RunE:                  CobraRunE,
		Args:                  cobra.NoArgs,
	}
	SourcePath   string
	BaselinePath string
	OutputFormat string
)

func init() {
	baseCmd.PersistentFlags().StringVar(&SourcePath, "source", "", "Source directory to scan")
	baseCmd.PersistentFlags().StringVar(&BaselinePath, "baseline", "", "Baseline directory to compare against")
	baseCmd.PersistentFlags().StringVar(&OutputFormat, "format", "table", "Output format: table|json")

	_ = baseCmd.MarkPersistentFlagRequired("source")
	_ = baseCmd.MarkPersistentFlagRequired("baseline")
}

func GetCommand() *cobra.Command {
	return baseCmd
}
