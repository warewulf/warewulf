package chmod

import (
	"testing"
)

// TestArgsOverlayChmod is a regression test for 260.
// Two arguments should fail, three should succeed.
func TestArgsOverlayChmod(t *testing.T) {
	command := GetCommand()

	err := command.Args(command, []string{"overlay_name", "file_name"})
	if err == nil {
		t.Errorf("two arguments to overlay chmod should fail")
	}

	err = command.Args(command, []string{"overlay_name", "file_name", "0755"})
	if err != nil {
		t.Errorf("three arguments to overlay chmod should succeed")
	}
}
