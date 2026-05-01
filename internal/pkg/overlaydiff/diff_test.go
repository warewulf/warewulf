package overlaydiff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDiff_IncludesAllExpectedChangeTypes verifies all supported change
// categories are reported correctly by Diff.
func TestDiff_IncludesAllExpectedChangeTypes(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	baselineDir := filepath.Join(tmpDir, "baseline")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}
	if !assert.NoError(t, os.MkdirAll(baselineDir, 0755)) {
		return
	}

	writeTestFile(t, filepath.Join(sourceDir, "added.txt"), "added")
	writeTestFile(t, filepath.Join(baselineDir, "removed.txt"), "removed")

	writeTestFile(t, filepath.Join(sourceDir, "modified-same-size.txt"), "xxxx")
	writeTestFile(t, filepath.Join(baselineDir, "modified-same-size.txt"), "yyyy")

	writeTestFile(t, filepath.Join(sourceDir, "mode-change.txt"), "same")
	writeTestFile(t, filepath.Join(baselineDir, "mode-change.txt"), "same")
	if !assert.NoError(t, os.Chmod(filepath.Join(sourceDir, "mode-change.txt"), 0644)) {
		return
	}
	if !assert.NoError(t, os.Chmod(filepath.Join(baselineDir, "mode-change.txt"), 0600)) {
		return
	}

	if !assert.NoError(t, os.Symlink("target-a", filepath.Join(sourceDir, "symlink-change"))) {
		return
	}
	if !assert.NoError(t, os.Symlink("target-b", filepath.Join(baselineDir, "symlink-change"))) {
		return
	}

	writeTestFile(t, filepath.Join(sourceDir, "type-change"), "file")
	if !assert.NoError(t, os.MkdirAll(filepath.Join(baselineDir, "type-change"), 0755)) {
		return
	}

	changes, err := Diff(sourceDir, baselineDir)
	if !assert.NoError(t, err) {
		return
	}

	changeByPath := make(map[string]Change)
	for _, change := range changes {
		changeByPath[change.Path] = change
	}

	assert.Equal(t, ChangeAdded, changeByPath["/added.txt"].Change)
	assert.Equal(t, ChangeRemoved, changeByPath["/removed.txt"].Change)
	assert.Equal(t, ChangeModified, changeByPath["/modified-same-size.txt"].Change)
	assert.Equal(t, ChangeModeChanged, changeByPath["/mode-change.txt"].Change)
	assert.Equal(t, ChangeModified, changeByPath["/symlink-change"].Change)
	assert.Equal(t, ChangeTypeChanged, changeByPath["/type-change"].Change)
}

// TestCompare_DeterministicOrder verifies Compare returns changes sorted by path.
func TestCompare_DeterministicOrder(t *testing.T) {
	source := map[string]Entry{
		"/b": {Path: "/b", Type: EntryFile, Hash: "1"},
		"/a": {Path: "/a", Type: EntryFile, Hash: "1"},
	}
	baseline := map[string]Entry{}

	changes := Compare(source, baseline)
	if !assert.Len(t, changes, 2) {
		return
	}
	assert.Equal(t, "/a", changes[0].Path)
	assert.Equal(t, "/b", changes[1].Path)
}

// TestFormatTableAndJSON verifies table and JSON formatting include key fields.
func TestFormatTableAndJSON(t *testing.T) {
	changes := []Change{{
		Path:   "/file.txt",
		Change: ChangeAdded,
		Type:   EntryFile,
		Mode:   0644,
		Size:   12,
	}}

	table := FormatTable(changes)
	assert.Contains(t, table, "CHANGE")
	assert.Contains(t, table, "added")
	assert.Contains(t, table, "/file.txt")

	jsonOut, err := FormatJSON(changes)
	if !assert.NoError(t, err) {
		return
	}
	assert.Contains(t, string(jsonOut), "\"change\": \"added\"")
	assert.Contains(t, string(jsonOut), "\"path\": \"/file.txt\"")
}

// writeTestFile creates parent directories and writes test content.
func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if !assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0755)) {
		return
	}
	assert.NoError(t, os.WriteFile(path, []byte(content), 0644))
}
