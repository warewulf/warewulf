package util

import (
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
)

// ConfirmationPrompt prompt is a blocking confirmation prompt.
// Returns true on y or yes user input.
func ConfirmationPrompt(label string) (yes bool) {

	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	result, _ := prompt.Run()
	if result == "y" || result == "yes" {
		yes = true
	}
	return
}

/*
Simple check if the config can be written in case wwctl isn't run as root
*/
func CanWriteConfig() (canwrite *wwapiv1.CanWriteConfig) {
	canwrite = new(wwapiv1.CanWriteConfig)
	err := syscall.Access(node.ConfigFile, syscall.O_RDWR)
	if err != nil {
		wwlog.Warn("Couldn't open %s:%s", node.ConfigFile, err)
		canwrite.CanWriteConfig = false
	} else {
		canwrite.CanWriteConfig = true
	}
	return canwrite
}
