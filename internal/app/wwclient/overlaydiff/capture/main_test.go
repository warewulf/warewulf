package capture

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaptureCommand_TableOutput(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	baselineDir := filepath.Join(tmpDir, "baseline")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}
	if !assert.NoError(t, os.MkdirAll(baselineDir, 0755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("new"), 0644)) {
		return
	}

	cmd := GetCommand()
	cmd.SetArgs([]string{"--source", sourceDir, "--baseline", baselineDir})
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, out.String(), "CHANGE")
	assert.Contains(t, out.String(), "added")
	assert.Contains(t, out.String(), "/file.txt")
}

func TestCaptureCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	baselineDir := filepath.Join(tmpDir, "baseline")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}
	if !assert.NoError(t, os.MkdirAll(baselineDir, 0755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("x"), 0644)) {
		return
	}

	cmd := GetCommand()
	cmd.SetArgs([]string{"--source", sourceDir, "--baseline", baselineDir, "--format", "json"})
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if !assert.NoError(t, err) {
		return
	}

	assert.Contains(t, out.String(), "\"change\": \"added\"")
	assert.Contains(t, out.String(), "\"path\": \"/new.txt\"")
}

func TestCaptureCommand_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	baselineDir := filepath.Join(tmpDir, "baseline")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}
	if !assert.NoError(t, os.MkdirAll(baselineDir, 0755)) {
		return
	}

	cmd := GetCommand()
	cmd.SetArgs([]string{"--source", sourceDir, "--baseline", baselineDir, "--format", "xml"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "expected table or json")
}
