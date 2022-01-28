package mkdir

import (
	"testing"
)

// TestArgsOverlayMkdir is a regression test for 248.
// One argument should fail, two should succeed.
func TestArgsOverlayMkdir(t *testing.T) {
	command := GetCommand()

	err := command.Args(command, []string{"overlay_name"})
	if err == nil {
		t.Errorf("one argument to overlay mkdir should fail")
	}

	err = command.Args(command, []string{"overlay_name", "directory"})
	if err != nil {
		t.Errorf("two arguments to overlay mkdir should succeed")
	}
}
