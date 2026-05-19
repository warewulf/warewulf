package capture

import (
	"bytes"
	"encoding/json"
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
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
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
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--format", "json"})
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

func TestStopCommand_ArtifactWritesOverlayStructureAndManifest(t *testing.T) {
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

	artifactRoot := filepath.Join(artifactParent, overlayName)
	rootfs := filepath.Join(artifactRoot, "rootfs")
	data, err := os.ReadFile(filepath.Join(rootfs, "a.txt"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "a-new", string(data))
	_, err = os.Stat(filepath.Join(rootfs, "b.txt"))
	assert.Error(t, err)

	manifest, err := overlaydiff.LoadArtifactManifest(filepath.Join(artifactRoot, overlaydiff.ArtifactManifestFileName))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, overlayName, manifest.OverlayName)
	assert.Equal(t, sourceDir, manifest.SourceRoot)
	assert.Equal(t, "node01", manifest.NodeSource)
	assert.Equal(t, []string{"/a.txt"}, manifest.SelectedPaths)
	assert.Equal(t, 1, manifest.Summary.Selected)

	assert.NoError(t, overlaydiff.ValidateArtifact(artifactRoot))
	assert.Contains(t, stopOut.String(), "Artifact exported 1 selected entries")
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
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--artifact"})
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
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--artifact", "--overlay-name", "demo", "--export"})
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
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
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
	stopCmd.SetArgs([]string{"--state-file", stateFile})
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.Contains(t, stopOut.String(), "modified")
	assert.Contains(t, stopOut.String(), "/file.txt")
}
