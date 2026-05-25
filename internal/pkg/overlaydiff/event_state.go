package overlaydiff

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

const eventStateVersion = 1

// EventHealth describes whether event-assisted state is suitable for fast-path use.
type EventHealth string

const (
	EventHealthOK       EventHealth = "ok"
	EventHealthDegraded EventHealth = "degraded"
)

// EventState stores event-assisted session metadata for start/stop workflows.
type EventState struct {
	Version     int         `json:"version"`
	Mode        string      `json:"mode"`
	SessionID   string      `json:"session_id"`
	SourceRoot  string      `json:"source_root"`
	IncludeRoots []string   `json:"include_roots,omitempty"`
	WatchRoots  []string    `json:"watch_roots,omitempty"`
	StartedAt   time.Time   `json:"started_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Health      EventHealth `json:"health"`
	Reasons     []string    `json:"reasons,omitempty"`
}

// DefaultEventStatePath returns the sidecar path for event-assisted state.
func DefaultEventStatePath(stateFile string) string {
	return stateFile + ".events.json"
}

// NewEventState builds initial event-assisted state.
func NewEventState(sourceRoot string, includeRoots []string) EventState {
	return EventState{
		Version:      eventStateVersion,
		Mode:         "event-assisted",
		SessionID:    uuid.NewString(),
		SourceRoot:   sourceRoot,
		IncludeRoots: append([]string(nil), includeRoots...),
		StartedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Health:       EventHealthOK,
	}
}

// SaveEventState writes event-assisted state atomically.
func SaveEventState(path string, state EventState) error {
	state.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create event state directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".event-state-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create event state temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write event state temp file: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to set event state temp file mode: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close event state temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to atomically replace event state: %w", err)
	}

	return nil
}

// LoadEventState loads event-assisted state from disk.
func LoadEventState(path string) (EventState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EventState{}, fmt.Errorf("failed to read event state: %w", err)
	}

	var state EventState
	if err := json.Unmarshal(data, &state); err != nil {
		return EventState{}, fmt.Errorf("failed to unmarshal event state: %w", err)
	}

	state.IncludeRoots = NormalizeExcludes(state.IncludeRoots)
	state.WatchRoots = NormalizeExcludes(state.WatchRoots)
	if state.Health == "" {
		state.Health = EventHealthDegraded
	}

	return state, nil
}

// ProbeEventWatchRoots validates watchability of candidate roots using inotify.
func ProbeEventWatchRoots(sourceRoot string, includeRoots []string) (watchRoots []string, reasons []string) {
	roots := includeRoots
	if len(roots) == 0 {
		roots = []string{"/"}
	}

	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC)
	if err != nil {
		return nil, []string{fmt.Sprintf("failed to initialize inotify watcher: %v", err)}
	}
	defer unix.Close(fd)

	watchable := make([]string, 0, len(roots))
	seen := make(map[string]struct{}, len(roots))
	for _, root := range roots {
		absRoot := filepath.Join(sourceRoot, strings.TrimPrefix(root, "/"))
		info, statErr := os.Stat(absRoot)
		if statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				continue
			}
			reasons = append(reasons, fmt.Sprintf("failed to inspect watch root %s: %v", root, statErr))
			continue
		}
		if !info.IsDir() {
			continue
		}

		norm := normalizeRelPath(root)
		if _, ok := seen[norm]; ok {
			continue
		}

		if _, addErr := unix.InotifyAddWatch(fd, absRoot, unix.IN_CREATE|unix.IN_DELETE|unix.IN_MODIFY|unix.IN_MOVED_FROM|unix.IN_MOVED_TO|unix.IN_ATTRIB); addErr != nil {
			reasons = append(reasons, fmt.Sprintf("failed to watch %s: %v", norm, addErr))
			continue
		}

		watchable = append(watchable, norm)
		seen[norm] = struct{}{}
	}

	sort.Strings(watchable)
	sort.Strings(reasons)
	return watchable, reasons
}
