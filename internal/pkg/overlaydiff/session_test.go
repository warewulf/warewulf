package overlaydiff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecisionState_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "capture.json")
	decisionPath := DefaultDecisionStatePath(stateFile)

	input := map[string]Decision{
		"/etc/a": DecisionYes,
		"/etc/b": DecisionNo,
	}
	if !assert.NoError(t, SaveDecisionState(decisionPath, input)) {
		return
	}

	loaded, err := LoadDecisionState(decisionPath)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 1, loaded.Version)
	assert.Equal(t, DecisionYes, loaded.Decisions["/etc/a"])
	assert.Equal(t, DecisionNo, loaded.Decisions["/etc/b"])
}

func TestLoadSnapshot_BackwardCompatibleDecisions(t *testing.T) {
	// Legacy snapshots without decisions should still load with defaults.
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "capture.json")

	legacy := `{
		"source_root": "/tmp/source",
		"created_at": "2026-04-01T00:00:00Z",
		"entries": {}
	}`

	if !assert.NoError(t, os.WriteFile(stateFile, []byte(legacy), 0o644)) {
		return
	}

	snapshot, err := LoadSnapshot(stateFile)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, snapshot.Decisions)
	assert.Empty(t, snapshot.Decisions)
}

func TestFilterChanges_ByTypeAndPrefix(t *testing.T) {
	// Prefix filtering should respect path boundaries (/etc should not match /etc2).
	changes := []Change{
		{Path: "/etc", Change: ChangeModified, Type: EntryDir},
		{Path: "/etc/a", Change: ChangeAdded, Type: EntryFile},
		{Path: "/etc/b", Change: ChangeModified, Type: EntryFile},
		{Path: "/etc2/b", Change: ChangeModified, Type: EntryFile},
		{Path: "/var/c", Change: ChangeModeChanged, Type: EntryFile},
	}

	filtered := FilterChanges(changes, FilterOptions{
		Only:       []ChangeType{ChangeModified, ChangeModeChanged},
		PathPrefix: []string{"/etc"},
	})

	if !assert.Len(t, filtered, 2) {
		return
	}
	assert.Equal(t, "/etc", filtered[0].Path)
	assert.Equal(t, "/etc/b", filtered[1].Path)
}

func TestParseChangeTypes_InvalidValue(t *testing.T) {
	// Unsupported --only values should return a descriptive validation error.
	_, err := ParseChangeTypes([]string{"removed"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected added|modified|mode-changed")
}
