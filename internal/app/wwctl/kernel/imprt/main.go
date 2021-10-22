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

	// Checking if container flag was set, then overwriting OptRoot
	kernelVersion := args[0]
	kernelName := kernelVersion
	if len(args) > 1 {
		kernelName = args[1]
	}
	if OptContainer != "" {
		if container.ValidSource(OptContainer) {
			OptRoot = container.RootFsDir(OptContainer)
		} else {
			wwlog.Printf(wwlog.ERROR, " %s is not a valid container", OptContainer)
			os.Exit(1)
		}
	}
	output, err := kernel.Build(kernelVersion, kernelName, OptRoot)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed building kernel: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("%s: %s\n", kernelName, output)
	}

	if SetDefault {

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

	return nil
}
