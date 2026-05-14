package capture

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

const defaultSourcePath = "/"

var (
	startSourcePath string
	startStateFile  string
	startExcludes   []string
)

func GetStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "start",
		Short:                 "Capture an overlay diff baseline",
		Long:                  "Capture a baseline snapshot for overlay diffing",
		RunE:                  runStart,
		Args:                  cobra.NoArgs,
	}

	cmd.Flags().StringVar(&startSourcePath, "source", "", "Source directory to scan (default: /)")
	cmd.Flags().StringVar(&startStateFile, "state-file", "", "Snapshot file path")
	cmd.Flags().StringArrayVar(&startExcludes, "exclude", nil, defaultExcludeHelp())
	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	sourcePath := strings.TrimSpace(startSourcePath)
	if sourcePath == "" {
		sourcePath = defaultSourcePath
	}

	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}

	excludes := overlaydiff.ResolveExcludes(startExcludes)
	entries, err := overlaydiff.ScanTreeWithOptions(sourceAbs, overlaydiff.ScanOptions{Excludes: excludes})
	if err != nil {
		return err
	}

	snapshot := overlaydiff.NewSnapshot(sourceAbs, excludes, entries)
	stateFile := resolveStateFilePath(startStateFile)
	if err := overlaydiff.SaveSnapshot(stateFile, snapshot); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Capture saved to %s\n", stateFile)
	return nil
}
