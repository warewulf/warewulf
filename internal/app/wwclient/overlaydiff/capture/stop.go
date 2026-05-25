package capture

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

var (
	stopSourcePath  string
	stopStateFile   string
	stopExcludes    []string
	stopFormat      string
	stopInteractive bool
	stopOnly        []string
	stopPathPrefix  []string
	stopExport      bool
	stopExportDir   string
	stopArtifact    bool
	stopArtifactDir string
	stopOverlayName string
	stopNodeSource  string
	stopEventAssisted bool
)

const decisionCheckpointInterval = 10

func GetStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "stop",
		Short:                 "Stop capture and show changes",
		Long:                  "Stop capture, compare against the baseline, and show changes",
		RunE:                  runStop,
		Args:                  cobra.NoArgs,
	}

	cmd.Flags().StringVar(&stopSourcePath, "source", "", "Source directory to scan (default: /; optimized)")
	cmd.Flags().StringVar(&stopStateFile, "state-file", "", "Snapshot file path")
	cmd.Flags().StringArrayVar(&stopExcludes, "exclude", nil, defaultExcludeHelp())
	cmd.Flags().StringVar(&stopFormat, "format", "table", "Output format: table|json")
	cmd.Flags().BoolVar(&stopInteractive, "interactive", false, "Prompt for per-path selection decisions")
	cmd.Flags().StringArrayVar(&stopOnly, "only", nil, "Only include change class (repeatable): added|modified|mode-changed")
	cmd.Flags().StringArrayVar(&stopPathPrefix, "path-prefix", nil, "Only include path prefix (repeatable, example: /etc)")
	cmd.Flags().BoolVar(&stopExport, "export", false, "Copy selected files to an export directory")
	cmd.Flags().StringVar(&stopExportDir, "export-dir", "", "Export destination directory (default: randomized /tmp/wwclient-overlaydiff-*)")
	cmd.Flags().BoolVar(&stopArtifact, "artifact", false, "Export selected files as an overlay artifact directory")
	cmd.Flags().StringVar(&stopArtifactDir, "artifact-dir", "", "Artifact parent directory (default: randomized /tmp/wwclient-overlay-artifact-*)")
	cmd.Flags().StringVar(&stopOverlayName, "overlay-name", "", "Overlay name for artifact mode")
	cmd.Flags().StringVar(&stopNodeSource, "node-source", "", "Optional node identifier stored in artifact metadata")
	cmd.Flags().BoolVar(&stopEventAssisted, "event-assisted", false, "Use event-assisted session metadata when available")

	return cmd
}

func runStop(cmd *cobra.Command, args []string) error {
	// Route human-readable status to stderr in JSON mode so stdout stays machine-parseable.
	format := strings.ToLower(stopFormat)
	if format != "table" && format != "json" {
		return fmt.Errorf("invalid format %q: expected table or json", stopFormat)
	}

	textOut := cmd.OutOrStdout()
	if format == "json" {
		textOut = cmd.ErrOrStderr()
	}

	onlyFilters, err := overlaydiff.ParseChangeTypes(stopOnly)
	if err != nil {
		return err
	}

	if stopArtifact && (stopExport || strings.TrimSpace(stopExportDir) != "") {
		return fmt.Errorf("--artifact mode can not be combined with --export/--export-dir")
	}
	if strings.TrimSpace(stopArtifactDir) != "" && !stopArtifact {
		return fmt.Errorf("--artifact-dir requires --artifact")
	}
	if strings.TrimSpace(stopOverlayName) != "" && !stopArtifact {
		return fmt.Errorf("--overlay-name requires --artifact")
	}
	if stopArtifact {
		if err := overlaydiff.ValidateOverlayName(stopOverlayName); err != nil {
			return err
		}
	}

	stateFile := resolveStateFilePath(stopStateFile)
	snapshot, err := overlaydiff.LoadSnapshot(stateFile)
	if err != nil {
		return err
	}
	decisionStatePath := overlaydiff.DefaultDecisionStatePath(stateFile)
	if decisionState, decisionErr := overlaydiff.LoadDecisionState(decisionStatePath); decisionErr == nil {
		snapshot.Decisions = decisionState.Decisions
	} else if !errors.Is(decisionErr, os.ErrNotExist) {
		return decisionErr
	}

	if stopEventAssisted {
		eventStatePath := overlaydiff.DefaultEventStatePath(stateFile)
		eventState, eventErr := overlaydiff.LoadEventState(eventStatePath)
		switch {
		case errors.Is(eventErr, os.ErrNotExist):
			_, _ = fmt.Fprintln(textOut, "Event-assisted state not found; falling back to full scan.")
		case eventErr != nil:
			_, _ = fmt.Fprintf(textOut, "Event-assisted state unreadable (%v); falling back to full scan.\n", eventErr)
		default:
			if eventState.SourceRoot != "" && snapshot.SourceRoot != "" && eventState.SourceRoot != snapshot.SourceRoot {
				_, _ = fmt.Fprintln(textOut, "Event-assisted source mismatch; falling back to full scan.")
			} else if eventState.Health != overlaydiff.EventHealthOK {
				reason := strings.Join(eventState.Reasons, "; ")
				if strings.TrimSpace(reason) == "" {
					reason = "degraded watcher health"
				}
				_, _ = fmt.Fprintf(textOut, "Event-assisted state degraded (%s); falling back to full scan.\n", reason)
			} else {
				_, _ = fmt.Fprintln(textOut, "Event-assisted session loaded; candidate journal fast-path is not enabled yet, using full scan.")
			}
		}
	}

	sourcePath := strings.TrimSpace(stopSourcePath)
	if sourcePath == "" {
		sourcePath = defaultSourcePath
	}

	sourceAbs, err := filepath.Abs(sourcePath)
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

	scanOptions := overlaydiff.ScanOptions{Excludes: excludes}
	if sourceAbs == defaultSourcePath {
		scanOptions.IncludeRoots = defaultIncludeRoots
	}
	scanOptions.BaselineEntries = snapshot.Entries
	entries, err := overlaydiff.ScanTreeWithOptions(sourceAbs, scanOptions)
	if err != nil {
		return err
	}

	changes := overlaydiff.Compare(entries, snapshot.Entries)
	changes = overlaydiff.FilterChanges(changes, overlaydiff.FilterOptions{
		Only:       onlyFilters,
		PathPrefix: stopPathPrefix,
	})

	snapshot.Changes = overlaydiff.SummarizeChanges(changes)
	if err := overlaydiff.SaveSnapshot(stateFile, snapshot); err != nil {
		return err
	}

	if stopInteractive {
		// Persist each answer immediately so interrupted sessions can resume safely.
		if err := runInteractiveSelection(cmd.InOrStdin(), textOut, changes, &snapshot, decisionStatePath); err != nil {
			if errors.Is(err, errInteractiveCancelled) {
				_, _ = fmt.Fprintln(textOut, "Cancelled. Decisions were saved and can be resumed later.")
				return err
			}
			return err
		}
	}

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

	selected, skipped, templated, unset := summarizeDecisions(changes, snapshot.Decisions)
	_, _ = fmt.Fprintf(textOut, "Decision summary: selected=%d skipped=%d templated=%d unset=%d\n", selected, skipped, templated, unset)

	if stopExport || strings.TrimSpace(stopExportDir) != "" {
		exportDir, err := prepareExportDir(strings.TrimSpace(stopExportDir))
		if err != nil {
			return err
		}

		exported, err := exportSelected(sourceAbs, exportDir, changes, snapshot.Decisions)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(textOut, "Exported %d selected entries to %s\n", exported, exportDir)
	}

	if stopArtifact {
		artifactParent, err := prepareArtifactParentDir(strings.TrimSpace(stopArtifactDir))
		if err != nil {
			return err
		}
		artifactRoot := filepath.Join(artifactParent, strings.TrimSpace(stopOverlayName))
		if err := ensureSecureDirPath(artifactParent, artifactRoot); err != nil {
			return err
		}

		artifactRootfs := filepath.Join(artifactRoot, "rootfs")
		if err := ensureSecureDirPath(artifactParent, artifactRootfs); err != nil {
			return err
		}

		exported, err := exportSelected(sourceAbs, artifactRootfs, changes, snapshot.Decisions)
		if err != nil {
			return err
		}

		manifest := overlaydiff.BuildArtifactManifest(
			strings.TrimSpace(stopOverlayName),
			sourceAbs,
			strings.TrimSpace(stopNodeSource),
			selectedPaths(changes, snapshot.Decisions),
			overlaydiff.DecisionSummary{
				Selected:  selected,
				Skipped:   skipped,
				Templated: templated,
				Unset:     unset,
			},
		)
		manifestPath := filepath.Join(artifactRoot, overlaydiff.ArtifactManifestFileName)
		if err := overlaydiff.SaveArtifactManifest(manifestPath, manifest); err != nil {
			return err
		}
		if err := overlaydiff.ValidateArtifact(artifactRoot); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(textOut, "Artifact exported %d selected entries to %s\n", exported, artifactRoot)
	}

	return nil
}

var errInteractiveCancelled = errors.New("interactive selection cancelled")

func runInteractiveSelection(in io.Reader, out io.Writer, changes []overlaydiff.Change, snapshot *overlaydiff.Snapshot, decisionStatePath string) error {
	reader := bufio.NewReader(in)
	pendingSaves := 0
	flush := func() error {
		if err := overlaydiff.SaveDecisionState(decisionStatePath, snapshot.Decisions); err != nil {
			return err
		}
		pendingSaves = 0
		return nil
	}

	for idx, change := range changes {
		currentDecision := normalizeDecision(snapshot.Decisions[change.Path])
		snapshot.Decisions[change.Path] = currentDecision

		if currentDecision != overlaydiff.DecisionUnset {
			continue
		}

		for {
			_, _ = fmt.Fprintf(out, "[%d/%d] %s %s (%s) -> (y)es, (n)o, (t)emplated, (e)xit: ", idx+1, len(changes), change.Change, change.Path, change.Type)
			answer, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					if flushErr := flush(); flushErr != nil {
						return flushErr
					}
					return errInteractiveCancelled
				}
				return err
			}

			switch strings.ToLower(strings.TrimSpace(answer)) {
			case "y", "yes":
				snapshot.Decisions[change.Path] = overlaydiff.DecisionYes
			case "n", "no":
				snapshot.Decisions[change.Path] = overlaydiff.DecisionNo
			case "t", "templated":
				snapshot.Decisions[change.Path] = overlaydiff.DecisionTemplated
			case "e", "exit":
				if err := flush(); err != nil {
					return err
				}
				return errInteractiveCancelled
			default:
				_, _ = fmt.Fprintln(out, "Invalid answer, use y/n/t/e")
				continue
			}

			pendingSaves++
			if pendingSaves >= decisionCheckpointInterval {
				if err := flush(); err != nil {
					return err
				}
			}
			break
		}
	}
	if err := flush(); err != nil {
		return err
	}

	return nil
}

func summarizeDecisions(changes []overlaydiff.Change, decisions map[string]overlaydiff.Decision) (selected int, skipped int, templated int, unset int) {
	for _, change := range changes {
		switch normalizeDecision(decisions[change.Path]) {
		case overlaydiff.DecisionYes:
			selected++
		case overlaydiff.DecisionNo:
			skipped++
		case overlaydiff.DecisionTemplated:
			templated++
		default:
			unset++
		}
	}
	return
}

// normalizeDecision treats unknown persisted values as unset to recover legacy/corrupt state.
func normalizeDecision(value overlaydiff.Decision) overlaydiff.Decision {
	switch value {
	case "", overlaydiff.DecisionUnset:
		return overlaydiff.DecisionUnset
	case overlaydiff.DecisionYes, overlaydiff.DecisionNo, overlaydiff.DecisionTemplated:
		return value
	default:
		return overlaydiff.DecisionUnset
	}
}

func prepareExportDir(custom string) (string, error) {
	if custom == "" {
		dir, err := os.MkdirTemp("/tmp", "wwclient-overlaydiff-")
		if err != nil {
			return "", fmt.Errorf("failed to create export directory in /tmp: %w", err)
		}
		return dir, nil
	}

	exportDir := filepath.Clean(custom)
	info, err := os.Lstat(exportDir)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("export directory must not be a symlink: %s", exportDir)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("export directory is not a directory: %s", exportDir)
		}
		return exportDir, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to inspect export directory %s: %w", exportDir, err)
	}

	absoluteExportDir, err := filepath.Abs(exportDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve export directory %s: %w", exportDir, err)
	}
	if err := ensureSecureDirPath(string(filepath.Separator), absoluteExportDir); err != nil {
		return "", err
	}
	return exportDir, nil
}

func prepareArtifactParentDir(custom string) (string, error) {
	if custom == "" {
		dir, err := os.MkdirTemp("/tmp", "wwclient-overlay-artifact-")
		if err != nil {
			return "", fmt.Errorf("failed to create artifact directory in /tmp: %w", err)
		}
		return dir, nil
	}

	artifactDir := filepath.Clean(custom)
	info, err := os.Lstat(artifactDir)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("artifact directory must not be a symlink: %s", artifactDir)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("artifact directory is not a directory: %s", artifactDir)
		}
		return artifactDir, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to inspect artifact directory %s: %w", artifactDir, err)
	}

	absoluteArtifactDir, err := filepath.Abs(artifactDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve artifact directory %s: %w", artifactDir, err)
	}
	if err := ensureSecureDirPath(string(filepath.Separator), absoluteArtifactDir); err != nil {
		return "", err
	}
	return artifactDir, nil
}

func selectedPaths(changes []overlaydiff.Change, decisions map[string]overlaydiff.Decision) []string {
	paths := make([]string, 0, len(changes))
	for _, change := range changes {
		if decisions[change.Path] != overlaydiff.DecisionYes {
			continue
		}
		if change.Source == nil {
			continue
		}
		paths = append(paths, change.Path)
	}
	return paths
}

// exportSelected writes only explicitly selected entries into the destination root.
func exportSelected(sourceRoot, exportDir string, changes []overlaydiff.Change, decisions map[string]overlaydiff.Decision) (int, error) {
	exported := 0
	for _, change := range changes {
		if decisions[change.Path] != overlaydiff.DecisionYes {
			continue
		}
		if change.Source == nil {
			continue
		}

		relPath := strings.TrimPrefix(change.Path, "/")
		sourcePath := filepath.Join(sourceRoot, relPath)
		destPath := filepath.Join(exportDir, relPath)
		if err := ensureWithinExportRoot(exportDir, destPath); err != nil {
			return exported, err
		}

		if err := ensureSecureDirPath(exportDir, filepath.Dir(destPath)); err != nil {
			return exported, err
		}

		if info, err := os.Lstat(destPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return exported, fmt.Errorf("refusing to write through symlinked destination: %s", destPath)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return exported, fmt.Errorf("failed to inspect export destination %s: %w", destPath, err)
		}

		switch change.Source.Type {
		case overlaydiff.EntryDir:
			if err := ensureSecureDirPath(exportDir, destPath); err != nil {
				return exported, err
			}
			if err := os.Chmod(destPath, fs.FileMode(change.Source.Mode)); err != nil {
				return exported, fmt.Errorf("failed to set mode for directory %s: %w", destPath, err)
			}
		case overlaydiff.EntrySymlink:
			_ = osRemoveAll(destPath)
			if err := osSymlink(change.Source.LinkTarget, destPath); err != nil {
				return exported, err
			}
		default:
			if err := copyFile(sourcePath, destPath, fs.FileMode(change.Source.Mode)); err != nil {
				return exported, err
			}
		}
		exported++
	}

	return exported, nil
}

// ensureWithinExportRoot rejects path traversal outside the target root.
func ensureWithinExportRoot(root string, target string) error {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return fmt.Errorf("failed to resolve export path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing to export outside target root: %s", target)
	}
	return nil
}

// ensureSecureDirPath creates/validates each parent directory while forbidding symlink hops.
func ensureSecureDirPath(root string, target string) error {
	if err := ensureWithinExportRoot(root, target); err != nil {
		return err
	}

	rel, err := filepath.Rel(root, target)
	if err != nil {
		return fmt.Errorf("failed to resolve export parent path: %w", err)
	}

	current := root
	if rel == "." {
		return nil
	}

	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to use symlinked directory: %s", current)
			}
			if !info.IsDir() {
				return fmt.Errorf("export path is not a directory: %s", current)
			}
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to inspect export directory %s: %w", current, err)
		}
		if err := osMkdirAll(current, 0o700); err != nil {
			return fmt.Errorf("failed to create export directory %s: %w", current, err)
		}
	}

	return nil
}

var (
	osMkdirAll  = os.MkdirAll
	osRemoveAll = os.RemoveAll
	osSymlink   = os.Symlink
)

func copyFile(sourcePath, destPath string, mode fs.FileMode) error {
	// Create destination with regular-file semantics, then restore source mode.
	input, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer input.Close()

	output, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create export file %s: %w", destPath, err)
	}

	if _, err = io.Copy(output, input); err != nil {
		_ = output.Close()
		return fmt.Errorf("failed to copy file %s to %s: %w", sourcePath, destPath, err)
	}

	if err = output.Close(); err != nil {
		return fmt.Errorf("failed to close export file %s: %w", destPath, err)
	}

	if err = os.Chmod(destPath, mode); err != nil {
		return fmt.Errorf("failed to set mode for %s: %w", destPath, err)
	}

	return nil
}
