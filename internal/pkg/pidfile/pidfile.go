// Package pidfile provides utilities for managing process ID files.
// Based on original work found here, github.com/soellman/pidfile
package pidfile

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	// ErrProcessRunning indicates that a process with the PID from the pidfile is currently running.
	ErrProcessRunning = errors.New("process is running")
	// ErrFileStale indicates that the pidfile exists but the process is not running.
	ErrFileStale = errors.New("pidfile exists but process is not running")
	// ErrFileInvalid indicates that the pidfile contains invalid or unparseable contents.
	ErrFileInvalid = errors.New("pidfile has invalid contents")
)

// Remove removes a pidfile from the filesystem.
// It returns an error if the file cannot be removed.
func Remove(filename string) error {
	return os.RemoveAll(filename)
}

// Write writes a pidfile containing the current process ID.
// It returns the PID and an error if the process is already running or pidfile is orphaned.
func Write(filename string) (int, error) {
	return WriteControl(filename, os.Getpid(), false)
}

// WriteControl writes a pidfile with the specified PID and control options.
// If overwrite is false and a stale pidfile exists, it returns ErrFileStale.
// If overwrite is true, it will overwrite stale pidfiles.
// It returns the PID and an error if the process is already running.
func WriteControl(filename string, pid int, overwrite bool) (int, error) {
	// Check for existing pid
	oldpid, err := pidfileContents(filename)
	if err != nil && !os.IsNotExist(err) {
		return oldpid, err
	}

	// We have a pid
	if err == nil {
		if pidIsRunning(oldpid) {
			return oldpid, ErrProcessRunning
		}
		if !overwrite {
			return -1, ErrFileStale
		}
	}

	// We're clear to (over)write the file
	return pid, os.WriteFile(filename, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

// pidfileContents reads and parses the PID from a pidfile.
// It returns the PID and an error if the file cannot be read or contains invalid data.
func pidfileContents(filename string) (int, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return 0, ErrFileInvalid
	}

	return pid, nil
}

// pidIsRunning checks if a process with the given PID is currently running.
// It returns true if the process is running, false otherwise.
func pidIsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	if err != nil && err.Error() == "no such process" {
		return false
	}

	if err != nil && err.Error() == "os: process already finished" {
		return false
	}

	return true
}
