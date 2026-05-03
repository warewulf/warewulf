package overlaydiff

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot captures a scan baseline for later comparison.
type Snapshot struct {
	SourceRoot string           `json:"source_root"`
	CreatedAt  time.Time        `json:"created_at"`
	Excludes   []string         `json:"excludes,omitempty"`
	Entries    map[string]Entry `json:"entries"`
}

// NewSnapshot constructs a snapshot for the given scan result.
func NewSnapshot(sourceRoot string, excludes []string, entries map[string]Entry) Snapshot {
	return Snapshot{
		SourceRoot: sourceRoot,
		CreatedAt:  time.Now().UTC(),
		Excludes:   NormalizeExcludes(excludes),
		Entries:    entries,
	}
}

// SaveSnapshot writes the snapshot as JSON to the given path.
func SaveSnapshot(path string, snapshot Snapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}
	return nil
}

// LoadSnapshot loads a snapshot from the given path.
func LoadSnapshot(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("failed to read snapshot: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	snapshot.Excludes = NormalizeExcludes(snapshot.Excludes)
	if snapshot.Entries == nil {
		snapshot.Entries = make(map[string]Entry)
	}

	return snapshot, nil
}
