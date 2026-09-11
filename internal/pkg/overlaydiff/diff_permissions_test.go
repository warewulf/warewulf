package overlaydiff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanTreeWithOptions_SkipsUnreadableFile(t *testing.T) {
	tmpDir := t.TempDir()
	if !assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ok.txt"), []byte("ok"), 0o644)) {
		return
	}

	blocked := filepath.Join(tmpDir, "blocked.txt")
	if !assert.NoError(t, os.WriteFile(blocked, []byte("blocked"), 0o600)) {
		return
	}
	if !assert.NoError(t, os.Chmod(blocked, 0o000)) {
		return
	}
	t.Cleanup(func() {
		_ = os.Chmod(blocked, 0o600)
	})

	entries, err := ScanTreeWithOptions(tmpDir, ScanOptions{})
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, entries, "/ok.txt")
	assert.NotContains(t, entries, "/blocked.txt")
}

func TestScanTreeWithOptions_SkipsUnreadableDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	if !assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ok.txt"), []byte("ok"), 0o644)) {
		return
	}

	blockedDir := filepath.Join(tmpDir, "blocked")
	if !assert.NoError(t, os.MkdirAll(blockedDir, 0o700)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(blockedDir, "inside.txt"), []byte("x"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.Chmod(blockedDir, 0o000)) {
		return
	}
	t.Cleanup(func() {
		_ = os.Chmod(blockedDir, 0o700)
	})

	entries, err := ScanTreeWithOptions(tmpDir, ScanOptions{})
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, entries, "/ok.txt")
	assert.NotContains(t, entries, "/blocked/inside.txt")
}
