package overlaydiff

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventState_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "capture.events.json")

	state := NewEventState("/", []string{"/etc", "/var/lib"})
	state.WatchRoots = []string{"/etc"}
	state.Reasons = []string{"sample reason"}
	state.Health = EventHealthDegraded

	if !assert.NoError(t, SaveEventState(statePath, state)) {
		return
	}

	loaded, err := LoadEventState(statePath)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, state.Mode, loaded.Mode)
	assert.Equal(t, state.SourceRoot, loaded.SourceRoot)
	assert.Equal(t, state.Health, loaded.Health)
	assert.Equal(t, []string{"/etc", "/var/lib"}, loaded.IncludeRoots)
	assert.Equal(t, []string{"/etc"}, loaded.WatchRoots)
}

func TestEventState_LoadRejectsUnsupportedVersion(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "capture.events.json")

	content := `{
  "version": 1,
  "mode": "event-assisted",
  "session_id": "legacy",
  "source_root": "/",
  "started_at": "2026-05-03T00:00:00Z",
  "updated_at": "2026-05-03T00:00:00Z",
  "health": "ok"
}`
	if !assert.NoError(t, os.WriteFile(statePath, []byte(content), 0o644)) {
		return
	}

	_, err := LoadEventState(statePath)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "unsupported event state version")
}

func TestEventState_LoadRejectsMissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "capture.events.json")

	state := EventState{
		Version:    eventStateVersion,
		Mode:       "event-assisted",
		SourceRoot: "/",
		StartedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Health:     EventHealthOK,
	}
	if !assert.NoError(t, SaveEventState(statePath, state)) {
		return
	}

	_, err := LoadEventState(statePath)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "missing session_id")
}

func TestProbeEventWatchRoots(t *testing.T) {
	tmpDir := t.TempDir()
	if !assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "etc"), 0o755)) {
		return
	}

	watchRoots, reasons := ProbeEventWatchRoots(tmpDir, []string{"/etc", "/missing"})
	assert.Contains(t, watchRoots, "/etc")
	assert.Empty(t, reasons)
}
