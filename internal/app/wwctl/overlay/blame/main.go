package blame

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

type blameLine struct {
	Path    string `json:"path"`
	Overlay string `json:"overlay"`
	Context string `json:"context"`
}

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		format := strings.ToLower(vars.Format)
		if format != "table" && format != "json" {
			return fmt.Errorf("invalid format %q: expected table or json", vars.Format)
		}

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %w", err)
		}

		nodeData, err := nodeDB.GetNode(args[0])
		if err != nil {
			return fmt.Errorf("could not get node %s: %w", args[0], err)
		}
		allNodes, err := nodeDB.FindAllNodes()
		if err != nil {
			return fmt.Errorf("could not get node list: %w", err)
		}

		prefix := normalizePathPrefix(vars.PathPrefix)
		var lines []blameLine

		contextLines, err := collectBlameLines(nodeData, allNodes, nodeData.SystemOverlay, "system", vars.ShowModeChanges, prefix)
		if err != nil {
			return err
		}
		lines = append(lines, contextLines...)

		contextLines, err = collectBlameLines(nodeData, allNodes, nodeData.RuntimeOverlay, "runtime", vars.ShowModeChanges, prefix)
		if err != nil {
			return err
		}
		lines = append(lines, contextLines...)

		return printBlameLines(cmd, lines, format)
	}
}

func printBlameLines(cmd *cobra.Command, lines []blameLine, format string) error {
	if format == "json" {
		data, err := json.MarshalIndent(lines, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", data)
		return err
	}

	pathWidth := 0
	overlayWidth := 0
	for _, line := range lines {
		pathWidth = max(pathWidth, len(line.Path))
		overlayWidth = max(overlayWidth, len(line.Overlay))
	}

	for _, line := range lines {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %-*s  [%s overlay]\n", pathWidth, line.Path, overlayWidth, line.Overlay, line.Context); err != nil {
			return err
		}
	}
	return nil
}

func collectBlameLines(nodeData node.Node, allNodes []node.Node, overlayNames []string, context string, includeDirs bool, prefix string) ([]blameLine, error) {
	var lines []blameLine
	for _, overlayName := range overlayNames {
		if !util.ValidString("^[a-zA-Z0-9-._:]+$", overlayName) {
			return nil, fmt.Errorf("overlay names contains illegal characters: %v", overlayNames)
		}

		overlayRoot, err := overlay.Get(overlayName)
		if err != nil {
			return nil, fmt.Errorf("could not get overlay %s: %w", overlayName, err)
		}

		overlayLines, err := collectOverlayLines(nodeData, allNodes, overlayRoot, overlayName, context, includeDirs, prefix)
		if err != nil {
			return nil, err
		}
		lines = append(lines, overlayLines...)
	}
	return lines, nil
}

func collectOverlayLines(nodeData node.Node, allNodes []node.Node, overlayRoot overlay.Overlay, overlayName string, context string, includeDirs bool, prefix string) ([]blameLine, error) {
	var lines []blameLine
	rootfs := overlayRoot.Rootfs()
	err := filepath.Walk(rootfs, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking overlay %s: %w", overlayName, err)
		}

		relPath, err := filepath.Rel(rootfs, walkPath)
		if err != nil {
			return fmt.Errorf("could not determine path for overlay %s: %w", overlayName, err)
		}
		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			if !includeDirs {
				return nil
			}
		} else if !isBlameFile(info) {
			return nil
		}

		deployedPath := deployedOverlayPath(relPath)
		if !info.IsDir() && filepath.Ext(walkPath) == ".ww" {
			deployedPath = strings.TrimSuffix(deployedPath, ".ww")
			paths, err := overlay.TemplateOutputPaths(walkPath, deployedPath, overlayName, nodeData, allNodes)
			if err != nil {
				return err
			}
			for _, templatePath := range paths {
				if !pathMatchesPrefix(templatePath, prefix) {
					continue
				}
				lines = append(lines, blameLine{
					Path:    templatePath,
					Overlay: overlayName,
					Context: context,
				})
			}
			return nil
		}
		if !pathMatchesPrefix(deployedPath, prefix) {
			return nil
		}

		lines = append(lines, blameLine{
			Path:    deployedPath,
			Overlay: overlayName,
			Context: context,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].Path < lines[j].Path
	})
	return lines, nil
}

func isBlameFile(info os.FileInfo) bool {
	return info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0
}

func deployedOverlayPath(relPath string) string {
	deployedPath := "/" + filepath.ToSlash(relPath)
	return path.Clean(deployedPath)
}

func normalizePathPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	prefix = path.Clean("/" + strings.TrimPrefix(filepath.ToSlash(prefix), "/"))
	return prefix
}

func pathMatchesPrefix(filePath string, prefix string) bool {
	if prefix == "" || prefix == "/" {
		return true
	}
	return filePath == prefix || strings.HasPrefix(filePath, prefix+"/")
}
