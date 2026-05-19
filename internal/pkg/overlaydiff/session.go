package overlaydiff

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Decision records interactive selection for a path.
type Decision string

const (
	// DecisionUnset indicates no explicit choice has been made yet.
	DecisionUnset Decision = "unset"
	// DecisionYes marks a path selected for export.
	DecisionYes Decision = "yes"
	// DecisionNo marks a path intentionally skipped.
	DecisionNo Decision = "no"
	// DecisionTemplated marks a path for template-based handling.
	DecisionTemplated Decision = "templated"
)

// ChangeSummary is persisted session metadata for a detected path.
type ChangeSummary struct {
	Path   string     `json:"path"`
	Change ChangeType `json:"change"`
	Type   EntryType  `json:"type"`
}

// SummarizeChanges converts detailed changes into a persistent summary.
func SummarizeChanges(changes []Change) []ChangeSummary {
	if len(changes) == 0 {
		return nil
	}

	result := make([]ChangeSummary, 0, len(changes))
	for _, change := range changes {
		result = append(result, ChangeSummary{Path: change.Path, Change: change.Change, Type: change.Type})
	}

	sort.Slice(result, func(i, j int) bool {
		// Keep ordering stable for persistence and predictable rendering.
		if result[i].Path == result[j].Path {
			return result[i].Change < result[j].Change
		}
		return result[i].Path < result[j].Path
	})

	return result
}

const decisionStateVersion = 1

// DecisionState stores interactive decisions outside of the main snapshot.
type DecisionState struct {
	Version   int                 `json:"version"`
	UpdatedAt time.Time           `json:"updated_at"`
	Decisions map[string]Decision `json:"decisions"`
}

// DefaultDecisionStatePath returns the sidecar path for a snapshot state file.
func DefaultDecisionStatePath(stateFile string) string {
	return stateFile + ".decisions.json"
}

// SaveDecisionState writes decisions atomically to the sidecar file.
func SaveDecisionState(path string, decisions map[string]Decision) error {
	state := DecisionState{
		Version:   decisionStateVersion,
		UpdatedAt: time.Now().UTC(),
		Decisions: make(map[string]Decision, len(decisions)),
	}
	for key, value := range decisions {
		state.Decisions[key] = value
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal decision state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create decision state directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".decision-state-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create decision state temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write decision state temp file: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to set decision state temp file mode: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close decision state temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to atomically replace decision state: %w", err)
	}

	return nil
}

// LoadDecisionState loads decisions from sidecar file.
func LoadDecisionState(path string) (DecisionState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DecisionState{}, fmt.Errorf("failed to read decision state: %w", err)
	}

	var state DecisionState
	if err := json.Unmarshal(data, &state); err != nil {
		return DecisionState{}, fmt.Errorf("failed to unmarshal decision state: %w", err)
	}
	if state.Decisions == nil {
		state.Decisions = make(map[string]Decision)
	}

	return state, nil
}
