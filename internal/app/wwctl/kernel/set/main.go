package set

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	if SetDefault {
		if len(args) != 1 {
			wwlog.Printf(wwlog.ERROR, "Can only set default for one kernel version\n")
		} else {
			nodeDB, err := node.New()
			if err != nil {
				return errors.Wrap(err, "Could not open node configuration")
			}

			//TODO: Don't loop through profiles, instead have a nodeDB function that goes directly to the map
			profiles, _ := nodeDB.FindAllProfiles()
			for _, profile := range profiles {
				wwlog.Printf(wwlog.DEBUG, "Looking for profile default: %s\n", profile.Id.Get())
				if profile.Id.Get() == "default" {
					wwlog.Printf(wwlog.DEBUG, "Found profile default, setting kernel version to: %s\n", args[0])
					profile.KernelVersion.Set(args[0])
					err := nodeDB.ProfileUpdate(profile)
					if err != nil {
						return errors.Wrap(err, "failed to update node profile")
					}
				}
			}
			err = nodeDB.Persist()
			if err != nil {
				return errors.Wrap(err, "failed to persist nodedb")
			}
			fmt.Printf("Set default kernel version to: %s\n", args[0])

			err = warewulfd.DaemonReload()
			if err != nil {
				return errors.Wrap(err, "failed to reload warewulf daemon")
			}
		}
	}

	return nil
}
