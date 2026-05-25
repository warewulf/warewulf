package capture

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

const defaultSourcePath = "/"

// defaultIncludeRoots keeps the default "/" workflow fast by restricting
// traversal to system-admin relevant trees.
var defaultIncludeRoots = []string{"/etc", "/var/lib", "/usr/local", "/opt", "/srv"}

var (
	startSourcePath string
	startStateFile  string
	startExcludes   []string
	startEventAssisted bool
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

	cmd.Flags().StringVar(&startSourcePath, "source", "", "Source directory to scan (default: / with optimized system roots)")
	cmd.Flags().StringVar(&startStateFile, "state-file", "", "Snapshot file path")
	cmd.Flags().StringArrayVar(&startExcludes, "exclude", nil, defaultExcludeHelp())
	cmd.Flags().BoolVar(&startEventAssisted, "event-assisted", false, "Enable event-assisted session metadata for faster stop runs")
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
	scanOptions := overlaydiff.ScanOptions{Excludes: excludes}
	if sourcePath == defaultSourcePath {
		// For default runs, scan only curated system roots.
		scanOptions.IncludeRoots = defaultIncludeRoots
	}
	entries, err := overlaydiff.ScanTreeWithOptions(sourceAbs, scanOptions)
	if err != nil {
		return err
	}

	snapshot := overlaydiff.NewSnapshot(sourceAbs, excludes, entries)
	stateFile := resolveStateFilePath(startStateFile)
	if err := overlaydiff.SaveSnapshot(stateFile, snapshot); err != nil {
		return err
	}

	if startEventAssisted {
		eventState := overlaydiff.NewEventState(sourceAbs, scanOptions.IncludeRoots)
		watchRoots, reasons := overlaydiff.ProbeEventWatchRoots(sourceAbs, scanOptions.IncludeRoots)
		eventState.WatchRoots = watchRoots
		eventState.Reasons = reasons
		if len(reasons) > 0 || len(watchRoots) == 0 {
			eventState.Health = overlaydiff.EventHealthDegraded
		}
		eventStatePath := overlaydiff.DefaultEventStatePath(stateFile)
		if err := overlaydiff.SaveEventState(eventStatePath, eventState); err != nil {
			return err
		}

		if eventState.Health == overlaydiff.EventHealthOK {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Event-assisted session initialized: %s\n", eventStatePath)
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Event-assisted session initialized in degraded mode: %s\n", eventStatePath)
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Capture saved to %s\n", stateFile)
	return nil
}
