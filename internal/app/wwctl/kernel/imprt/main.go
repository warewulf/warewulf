package imprt

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		// Checking if container flag was set, then overwriting OptRoot
		if OptContainer != "" {
			if container.ValidSource(OptContainer) == true {
				OptRoot = container.RootFsDir(OptContainer)
			} else {
				wwlog.Printf(wwlog.ERROR, " %s is not a valid container", OptContainer)
				os.Exit(1)
			}
		}
		output, err := kernel.Build(arg, OptRoot)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Failed building kernel: %s\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("%s: %s\n", arg, output)
		}
	}

	if SetDefault {
		if len(args) != 1 {
			wwlog.Printf(wwlog.ERROR, "Can only set default for one kernel version\n")
		} else {
			nodeDB, err := node.New()
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
				os.Exit(1)
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
