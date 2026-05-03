package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

var (
	stopSourcePath string
	stopStateFile  string
	stopExcludes   []string
	stopFormat     string
)

func GetStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "stop",
		Short:                 "Stop capture and show changes",
		Long:                  "Stop capture, compare against the baseline, and show changes",
		RunE:                  runStop,
		Args:                  cobra.NoArgs,
	}

	cmd.Flags().StringVar(&stopSourcePath, "source", "", "Source directory to scan")
	cmd.Flags().StringVar(&stopStateFile, "state-file", "", "Snapshot file path")
	cmd.Flags().StringArrayVar(&stopExcludes, "exclude", nil, defaultExcludeHelp())
	cmd.Flags().StringVar(&stopFormat, "format", "table", "Output format: table|json")

	_ = cmd.MarkFlagRequired("source")
	return cmd
}

func runStop(cmd *cobra.Command, args []string) error {
	format := strings.ToLower(stopFormat)
	if format != "table" && format != "json" {
		return fmt.Errorf("invalid format %q: expected table or json", stopFormat)
	}

	stateFile := resolveStateFilePath(stopStateFile)
	snapshot, err := overlaydiff.LoadSnapshot(stateFile)
	if err != nil {
		return err
	}

	sourceAbs, err := filepath.Abs(stopSourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}
	if snapshot.SourceRoot != "" && sourceAbs != snapshot.SourceRoot {
		return fmt.Errorf("source path %s does not match capture %s", sourceAbs, snapshot.SourceRoot)
	}

	var excludes []string
	if len(stopExcludes) > 0 {
		excludes = overlaydiff.ResolveExcludes(stopExcludes)
	} else if len(snapshot.Excludes) > 0 {
		excludes = overlaydiff.NormalizeExcludes(snapshot.Excludes)
	} else {
		excludes = overlaydiff.ResolveExcludes(nil)
	}

	entries, err := overlaydiff.ScanTreeWithOptions(sourceAbs, overlaydiff.ScanOptions{Excludes: excludes})
	if err != nil {
		return err
	}

	changes := overlaydiff.Compare(entries, snapshot.Entries)
	switch format {
	case "json":
		data, err := overlaydiff.FormatJSON(changes)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", data)
	default:
		_, _ = fmt.Fprint(cmd.OutOrStdout(), overlaydiff.FormatTable(changes))
	}

	if err := os.Remove(stateFile); err != nil {
		return fmt.Errorf("failed to remove snapshot: %w", err)
	}
	return nil
}
