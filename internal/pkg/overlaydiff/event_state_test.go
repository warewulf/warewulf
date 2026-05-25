package overlaydiff

import (
	"os"
	"path/filepath"
	"testing"

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

func TestProbeEventWatchRoots(t *testing.T) {
	tmpDir := t.TempDir()
	if !assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "etc"), 0o755)) {
		return
	}

	watchRoots, reasons := ProbeEventWatchRoots(tmpDir, []string{"/etc", "/missing"})
	assert.Contains(t, watchRoots, "/etc")
	assert.Empty(t, reasons)
}
