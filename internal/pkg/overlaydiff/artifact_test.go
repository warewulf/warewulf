package overlaydiff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOverlayName(t *testing.T) {
	// Accept safe names and reject whitespace/path-traversal variants.
	assert.NoError(t, ValidateOverlayName("demo-overlay_1.0"))
	assert.Error(t, ValidateOverlayName(""))
	assert.Error(t, ValidateOverlayName("demo overlay"))
	assert.Error(t, ValidateOverlayName("../../escape"))
}

func TestValidateArtifact_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	artifactRoot := filepath.Join(tmpDir, "demo")
	rootfs := filepath.Join(artifactRoot, "rootfs")

	if !assert.NoError(t, os.MkdirAll(filepath.Join(rootfs, "etc"), 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(rootfs, "etc", "config"), []byte("x"), 0o644)) {
		return
	}

	// Build a minimal valid artifact layout: rootfs payload + manifest.
	manifest := BuildArtifactManifest("demo", "/tmp/source", "node-a", []string{"/etc/config"}, DecisionSummary{Selected: 1})
	if !assert.NoError(t, SaveArtifactManifest(filepath.Join(artifactRoot, ArtifactManifestFileName), manifest)) {
		return
	}

	assert.NoError(t, ValidateArtifact(artifactRoot))
}

func TestValidateArtifact_InvalidManifestPath(t *testing.T) {
	tmpDir := t.TempDir()
	artifactRoot := filepath.Join(tmpDir, "demo")
	rootfs := filepath.Join(artifactRoot, "rootfs")

	if !assert.NoError(t, os.MkdirAll(rootfs, 0o755)) {
		return
	}

	// Non-normalized path should fail validation.
	manifest := ArtifactManifest{
		SchemaVersion: "v1",
		OverlayName:   "demo",
		SourceRoot:    "/tmp/source",
		SelectedPaths: []string{"etc/config"},
		Summary:       DecisionSummary{Selected: 1},
	}
	if !assert.NoError(t, SaveArtifactManifest(filepath.Join(artifactRoot, ArtifactManifestFileName), manifest)) {
		return
	}

	err := ValidateArtifact(artifactRoot)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "normalized absolute-style")
}
