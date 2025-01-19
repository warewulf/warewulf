package build

import (
	"testing"
)

// TestArgsImageBuild is a regression test for 215.
func TestArgsImageBuild(t *testing.T) {
	command := GetCommand()

	err := command.Args(command, []string{})
	if err != nil {
		t.Errorf("no arguments to image build should succeed.")
	}

	err = command.Args(command, []string{"image1", "image2"})
	if err != nil {
		t.Errorf("multiple arguments to image build should succeed.")
	}
}
