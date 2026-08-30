package capture

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

func TestStartStopCommand_TableOutput(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("old"), 0644)) {
		return
	}

	// Capture baseline, mutate source, then verify table output and summary.
	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startOut := new(bytes.Buffer)
	startCmd.SetOut(startOut)
	startCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("new"), 0644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--no-interactive"})
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.Contains(t, stopOut.String(), "CHANGE")
	assert.Contains(t, stopOut.String(), "modified")
	assert.Contains(t, stopOut.String(), "/file.txt")
	assert.Contains(t, stopOut.String(), "Decision summary:")

	_, err := os.Stat(stateFile)
	assert.NoError(t, err)
}

func TestStartStopCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("x"), 0644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("y"), 0644)) {
		return
	}

	// JSON payload should remain clean; summary goes to stderr.
	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--format", "json", "--no-interactive"})
	stopOut := new(bytes.Buffer)
	stopErr := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(stopErr)

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.NotContains(t, stopOut.String(), "Decision summary:")

	var payload []overlaydiff.Change
	if !assert.NoError(t, json.Unmarshal(stopOut.Bytes(), &payload)) {
		return
	}
	if !assert.NotEmpty(t, payload) {
		return
	}

	assert.Equal(t, overlaydiff.ChangeModified, payload[0].Change)
	assert.Equal(t, "/new.txt", payload[0].Path)

	assert.Contains(t, stopErr.String(), "Decision summary:")
}

func TestStopCommand_DefaultsToInteractive(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.Contains(t, stopOut.String(), "-> (y)es, (n)o, (e)xit")
}

func TestStopCommand_InteractivePersistsDecision(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new"), 0o644)) {
		return
	}

	// Answering once should persist a stable decision in sidecar state.
	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive"})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	decisionStateFile := overlaydiff.DefaultDecisionStatePath(stateFile)
	data, err := os.ReadFile(decisionStateFile)
	if !assert.NoError(t, err) {
		return
	}
	assert.Contains(t, string(data), "\"/new.txt\": \"yes\"")
}

func TestStopCommand_InteractiveRecoversInvalidPersistedDecision(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new"), 0o644)) {
		return
	}

	// Simulate a legacy/invalid stored decision and verify re-selection recovery.
	snapshot, err := overlaydiff.LoadSnapshot(stateFile)
	if !assert.NoError(t, err) {
		return
	}
	snapshot.Decisions["/new.txt"] = overlaydiff.Decision("maybe")
	if !assert.NoError(t, overlaydiff.SaveSnapshot(stateFile, snapshot)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive"})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	decisionState, err := overlaydiff.LoadDecisionState(overlaydiff.DefaultDecisionStatePath(stateFile))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, overlaydiff.DecisionYes, decisionState.Decisions["/new.txt"])
}

func TestStopCommand_InteractiveUsesSnapshotDecisionsFallback(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new"), 0o644)) {
		return
	}

	snapshot, err := overlaydiff.LoadSnapshot(stateFile)
	if !assert.NoError(t, err) {
		return
	}
	snapshot.Decisions["/new.txt"] = overlaydiff.DecisionYes
	if !assert.NoError(t, overlaydiff.SaveSnapshot(stateFile, snapshot)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive"})
	stopCmd.SetIn(strings.NewReader(""))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	decisionState, err := overlaydiff.LoadDecisionState(overlaydiff.DefaultDecisionStatePath(stateFile))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, overlaydiff.DecisionYes, decisionState.Decisions["/new.txt"])
}

func TestStopCommand_FilterAndExportSelected(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	exportDir := filepath.Join(tmpDir, "export")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("a-old"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "b.txt"), []byte("b-old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("a-new"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "b.txt"), []byte("b-new"), 0o644)) {
		return
	}

	// Filter to one path and export only explicit selections.
	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--only", "modified", "--path-prefix", "/a.txt", "--export", "--export-dir", exportDir})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	data, err := os.ReadFile(filepath.Join(exportDir, "a.txt"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "a-new", string(data))

	_, err = os.Stat(filepath.Join(exportDir, "b.txt"))
	assert.Error(t, err)
}

func TestStopCommand_ExportDirectoryPreservesMode(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	exportDir := filepath.Join(tmpDir, "export")
	dirPath := filepath.Join(sourceDir, "dir")

	if !assert.NoError(t, os.MkdirAll(dirPath, 0o755)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.Chmod(dirPath, 0o3750)) {
		return
	}
	sourceInfo, err := os.Stat(dirPath)
	if !assert.NoError(t, err) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--only", "mode-changed", "--path-prefix", "/dir", "--export", "--export-dir", exportDir})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	info, err := os.Stat(filepath.Join(exportDir, "dir"))
	if !assert.NoError(t, err) {
		return
	}

	modeMask := os.ModePerm | os.ModeSetgid | os.ModeSticky
	assert.Equal(t, sourceInfo.Mode()&modeMask, info.Mode()&modeMask)
}

func TestStopCommand_DefaultExportDirIsPrivateAndRandomized(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--export"})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	output := stopOut.String()
	prefix := "Exported 1 selected entries to "
	idx := strings.Index(output, prefix)
	if !assert.NotEqual(t, -1, idx) {
		return
	}
	line := output[idx+len(prefix):]
	line = strings.SplitN(line, "\n", 2)[0]
	exportDir := strings.TrimSpace(line)

	assert.True(t, strings.HasPrefix(exportDir, "/tmp/wwclient-overlaydiff-"))

	info, err := os.Stat(exportDir)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}

func TestStopCommand_RejectsSymlinkExportDir(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	targetDir := filepath.Join(tmpDir, "target")
	symlinkDir := filepath.Join(tmpDir, "symlink")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.MkdirAll(targetDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("old"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.Symlink(targetDir, symlinkDir)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--export", "--export-dir", symlinkDir})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	err := stopCmd.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "must not be a symlink")
}

func TestStopCommand_RejectsSymlinkAncestorInExportDir(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	targetParent := filepath.Join(tmpDir, "target-parent")
	symlinkParent := filepath.Join(tmpDir, "symlink-parent")
	exportDir := filepath.Join(symlinkParent, "nested")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.MkdirAll(targetParent, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("old"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.Symlink(targetParent, symlinkParent)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--export", "--export-dir", exportDir})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	err := stopCmd.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "refusing to use symlinked directory")
}

func TestStopCommand_RejectsNonDirectoryExportDir(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	exportFile := filepath.Join(tmpDir, "export-file")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("old"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(exportFile, []byte("x"), 0o600)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--export", "--export-dir", exportFile})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	err := stopCmd.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "not a directory")
}

func TestStopCommand_ArtifactWritesArchiveWithOverlayStructureAndManifest(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	artifactParent := filepath.Join(tmpDir, "artifacts")
	overlayName := "demo-overlay"

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("a-old"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "b.txt"), []byte("b-old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("a-new"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "b.txt"), []byte("b-new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--only", "modified", "--path-prefix", "/a.txt", "--artifact", "--artifact-dir", artifactParent, "--overlay-name", overlayName, "--node-source", "node01"})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	artifactPath := filepath.Join(artifactParent, overlayName+".tar.gz")
	entries, err := readArtifactArchiveEntries(artifactPath)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "a-new", entries["rootfs/a.txt"])
	_, ok := entries["rootfs/b.txt"]
	assert.False(t, ok)

	manifest, err := loadManifestFromArchive(artifactPath)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, overlayName, manifest.OverlayName)
	assert.Equal(t, sourceDir, manifest.SourceRoot)
	assert.Equal(t, "node01", manifest.NodeSource)
	assert.Equal(t, []string{"/a.txt"}, manifest.SelectedPaths)
	assert.Equal(t, 1, manifest.Summary.Selected)

	assert.FileExists(t, artifactPath)
	assert.Contains(t, stopOut.String(), "Artifact exported 1 selected entries")
	assert.Contains(t, stopOut.String(), artifactPath)
}

func readArtifactArchiveEntries(archivePath string) (map[string]string, error) {
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer archiveFile.Close()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	entries := make(map[string]string)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, err
		}
		entries[header.Name] = string(data)
	}
	return entries, nil
}

func loadManifestFromArchive(archivePath string) (overlaydiff.ArtifactManifest, error) {
	entries, err := readArtifactArchiveEntries(archivePath)
	if err != nil {
		return overlaydiff.ArtifactManifest{}, err
	}
	manifestFile, err := os.CreateTemp("", "overlaydiff-manifest-test-*.json")
	if err != nil {
		return overlaydiff.ArtifactManifest{}, err
	}
	manifestPath := manifestFile.Name()
	if err := manifestFile.Close(); err != nil {
		return overlaydiff.ArtifactManifest{}, err
	}
	if err := os.WriteFile(manifestPath, []byte(entries[overlaydiff.ArtifactManifestFileName]), 0o600); err != nil {
		return overlaydiff.ArtifactManifest{}, err
	}
	defer os.Remove(manifestPath)
	return overlaydiff.LoadArtifactManifest(manifestPath)
}

func TestStopCommand_ArtifactRequiresOverlayName(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--artifact", "--no-interactive"})
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	err := stopCmd.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "overlay name")
}

func TestStopCommand_ArtifactAndExportMutuallyExclusive(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--artifact", "--overlay-name", "demo", "--export", "--no-interactive"})
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	err := stopCmd.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "can not be combined")
}

func TestStopCommand_MissingSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "missing.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--no-interactive"})
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	assert.Error(t, stopCmd.Execute())
}

func TestStartStopCommand_SourceFlagOptional(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startSourcePath = sourceDir
	startCmd.SetArgs([]string{"--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopSourcePath = sourceDir
	stopCmd.SetArgs([]string{"--state-file", stateFile, "--no-interactive"})
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.Contains(t, stopOut.String(), "modified")
	assert.Contains(t, stopOut.String(), "/file.txt")
}
