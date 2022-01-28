package build

import (
	"testing"
)

// TestArgsContainerBuild is a regression test for 215.
func TestArgsContainerBuild(t *testing.T) {
	command := GetCommand()

	err := command.Args(command, []string{})
	if err != nil {
		t.Errorf("no arguments to container build should succeed.")
	}

	err = command.Args(command, []string{"container1", "container2"})
	if err != nil {
		t.Errorf("multiple arguments to container build should succeed.")
	}
}
