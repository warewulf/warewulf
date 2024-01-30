package util

import (
	"path"
	"syscall"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"github.com/manifoldco/promptui"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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
	conf := warewulfconf.Get()
	confFile := path.Join(conf.Paths.Sysconfdir, "warewulf/nodes.conf")
	err := syscall.Access(confFile, syscall.O_RDWR)
	if err != nil {
		wwlog.Warn("Couldn't open %s:%s", confFile, err)
		canwrite.CanWriteConfig = false
	} else {
		canwrite.CanWriteConfig = true
	}
	return canwrite
}
