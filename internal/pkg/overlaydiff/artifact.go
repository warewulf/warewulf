package overlaydiff

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	// ArtifactManifestFileName is the JSON manifest stored at artifact root.
	ArtifactManifestFileName = "overlaydiff-manifest.json"
	// ArtifactSchemaVersion is the supported artifact manifest schema.
	ArtifactSchemaVersion = "v1"
)

var overlayNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// DecisionSummary captures persisted counts for operator decisions.
type DecisionSummary struct {
	Selected int `json:"selected"`
	Skipped  int `json:"skipped"`
	Unset    int `json:"unset"`
}

// ArtifactManifest stores metadata for exported overlaydiff artifacts.
type ArtifactManifest struct {
	SchemaVersion string          `json:"schema_version"`
	OverlayName   string          `json:"overlay_name"`
	CreatedAt     time.Time       `json:"created_at"`
	SourceRoot    string          `json:"source_root"`
	NodeSource    string          `json:"node_source,omitempty"`
	SelectedPaths []string        `json:"selected_paths"`
	Summary       DecisionSummary `json:"summary"`
}

// ValidateOverlayName enforces a safe overlay directory name.
func ValidateOverlayName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("overlay name must not be empty")
	}
	if !overlayNamePattern.MatchString(trimmed) {
		return fmt.Errorf("invalid overlay name %q: use only letters, numbers, dot, dash, underscore", name)
	}
	return nil
}

// BuildArtifactManifest creates a normalized v1 manifest payload.
func BuildArtifactManifest(overlayName, sourceRoot, nodeSource string, selectedPaths []string, summary DecisionSummary) ArtifactManifest {
	normalized := make([]string, 0, len(selectedPaths))
	seen := make(map[string]struct{}, len(selectedPaths))
	for _, value := range selectedPaths {
		normalizedPath := normalizeRelPath(strings.TrimSpace(value))
		if _, ok := seen[normalizedPath]; ok {
			continue
		}
		seen[normalizedPath] = struct{}{}
		normalized = append(normalized, normalizedPath)
	}
	sort.Strings(normalized)

	return ArtifactManifest{
		SchemaVersion: ArtifactSchemaVersion,
		OverlayName:   strings.TrimSpace(overlayName),
		CreatedAt:     time.Now().UTC(),
		SourceRoot:    sourceRoot,
		NodeSource:    strings.TrimSpace(nodeSource),
		SelectedPaths: normalized,
		Summary:       summary,
	}
}

// SaveArtifactManifest writes the artifact manifest JSON to disk.
func SaveArtifactManifest(path string, manifest ArtifactManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal artifact manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write artifact manifest: %w", err)
	}

	return nil
}

// LoadArtifactManifest reads and parses a manifest JSON file.
func LoadArtifactManifest(path string) (ArtifactManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ArtifactManifest{}, fmt.Errorf("failed to read artifact manifest: %w", err)
	}

	var manifest ArtifactManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return ArtifactManifest{}, fmt.Errorf("failed to parse artifact manifest: %w", err)
	}

	return manifest, nil
}

// ValidateArtifact performs structural and path-safety checks for an artifact.
func ValidateArtifact(root string) error {
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("failed to inspect artifact root %s: %w", root, err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("artifact root must not be a symlink: %s", root)
	}
	if !rootInfo.IsDir() {
		return fmt.Errorf("artifact root is not a directory: %s", root)
	}

	manifestPath := filepath.Join(root, ArtifactManifestFileName)
	manifest, err := LoadArtifactManifest(manifestPath)
	if err != nil {
		return err
	}
	if manifest.SchemaVersion != ArtifactSchemaVersion {
		return fmt.Errorf("unsupported artifact schema version %q: expected %s", manifest.SchemaVersion, ArtifactSchemaVersion)
	}

	if err := ValidateOverlayName(manifest.OverlayName); err != nil {
		return err
	}

	rootfs := filepath.Join(root, "rootfs")
	rootfsInfo, err := os.Lstat(rootfs)
	if err != nil {
		return fmt.Errorf("failed to inspect artifact rootfs %s: %w", rootfs, err)
	}
	if rootfsInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("artifact rootfs must not be a symlink: %s", rootfs)
	}
	if !rootfsInfo.IsDir() {
		return fmt.Errorf("artifact rootfs is not a directory: %s", rootfs)
	}

	for _, selectedPath := range manifest.SelectedPaths {
		normalized := normalizeRelPath(strings.TrimSpace(selectedPath))
		if normalized != selectedPath {
			return fmt.Errorf("manifest path must be normalized absolute-style: %s", selectedPath)
		}

		rel := strings.TrimPrefix(normalized, "/")
		target := filepath.Join(rootfs, rel)
		relToRootfs, err := filepath.Rel(rootfs, target)
		if err != nil {
			return fmt.Errorf("failed to resolve manifest path %s: %w", selectedPath, err)
		}
		if relToRootfs == ".." || strings.HasPrefix(relToRootfs, ".."+string(filepath.Separator)) {
			return fmt.Errorf("manifest path escapes rootfs: %s", selectedPath)
		}

		if _, err := os.Lstat(target); err != nil {
			return fmt.Errorf("manifest path does not exist in artifact: %s", selectedPath)
		}
	}

	return nil
}
