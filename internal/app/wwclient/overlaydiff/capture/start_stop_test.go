package capture

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

	_, err := os.Stat(stateFile)
	assert.Error(t, err)
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

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--format", "json"})
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.Contains(t, stopOut.String(), "\"change\": \"modified\"")
	assert.Contains(t, stopOut.String(), "\"path\": \"/new.txt\"")
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
