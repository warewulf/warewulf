package imprt

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	// Shim in a name if none given.
	name := ""
	if len(args) == 2 {
		name = args[1]
	}

	err = container.Import(&container.ImportParameter{
		Source:   args[0],
		Name:     name,
		Force:    SetForce,
		Update:   SetUpdate,
		Build:    SetBuild,
		Default:  SetDefault,
		SyncUser: SyncUser,
	})

	if err != nil {
		return
	}

	// we need to reload the daemon to reflect profile container changes
	if SetDefault {
		err = warewulfd.DaemonStatus()
		if err != nil {
			// warewulfd is not running, skip
			return nil
		}
		return warewulfd.DaemonReload()
	}

	return
}
