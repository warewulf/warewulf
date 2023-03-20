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
	if len(args) == 0 && !OptDetect {
		wwlog.Error("the '--detect' flag is needed, if no kernel version is suppiled")
		os.Exit(1)
	}
	if OptDetect && (OptRoot == "" || OptContainer == "") {
		wwlog.Error("the '--detect flag needs the '--container' or '--root' flag")
		os.Exit(1)
	}
	// Checking if container flag was set, then overwriting OptRoot
	if OptContainer != "" {
		if container.ValidSource(OptContainer) {
			OptRoot = container.RootFsDir(OptContainer)
		} else {
			wwlog.Error(" %s is not a valid container", OptContainer)
			os.Exit(1)
		}
	}

	var kernelVersion string
	var err error
	if len(args) > 0 {
		kernelVersion = args[0]
	} else {
		kernelVersion, err = kernel.FindKernelVersion(OptRoot)
		if err != nil {
			wwlog.Error("could not detect kernel under %s", OptRoot)
			os.Exit(1)
		}
	}
	kernelName := kernelVersion
	if len(args) > 1 {
		kernelName = args[1]
	} else if OptDetect && (OptContainer != "") {
		kernelName = OptContainer
	}
	output, err := kernel.Build(kernelVersion, kernelName, OptRoot)
	if err != nil {
		wwlog.Error("Failed building kernel: %s", err)
		os.Exit(1)
	} else {
		fmt.Printf("%s: %s\n", kernelName, output)
	}

	if SetDefault {

		nodeDB, err := node.ReadNodeYaml()
		if err != nil {
			wwlog.Error("Could not open node configuration: %s", err)
			os.Exit(1)
		}
		//TODO: Don't loop through profiles, instead have a nodeDB function that goes directly to the map
		profiles, _ := nodeDB.FindAllProfiles()
		for _, profile := range profiles {
			wwlog.Debug("Looking for profile default: %s", profile.Id.Get())
			if profile.Id.Get() == "default" {
				wwlog.Debug("Found profile default, setting kernel version to: %s", args[0])
				profile.Kernel.Override.Set(args[0])
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
