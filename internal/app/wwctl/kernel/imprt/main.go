package imprt

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && !OptDetect {
		return fmt.Errorf("the '--detect' flag is needed, if no kernel version is suppiled")
	}
	if OptDetect && (OptRoot == "" || OptContainer == "") {
		return fmt.Errorf("the '--detect flag needs the '--container' or '--root' flag")
	}
	// Checking if container flag was set, then overwriting OptRoot
	if OptContainer != "" {
		if container.ValidSource(OptContainer) {
			OptRoot = container.RootFsDir(OptContainer)
		} else {
			return fmt.Errorf(" %s is not a valid container", OptContainer)
		}
	}

	var kernelVersion string
	var err error
	if len(args) > 0 {
		kernelVersion = args[0]
	} else {
		_, kernelVersion, err = kernel.FindKernel(OptRoot)
		if err != nil {
			return err
		}
	}
	kernelName := kernelVersion
	if len(args) > 1 {
		kernelName = args[1]
	} else if OptDetect && (OptContainer != "") {
		kernelName = OptContainer
	}
	err = kernel.Build(kernelVersion, kernelName, OptRoot)
	if err != nil {
		return fmt.Errorf("failed building kernel: %s", err)
	} else {
		fmt.Printf("%s: %s\n", kernelName, "Finished kernel build")
	}

	if SetDefault {

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %s", err)
		}
		//TODO: Don't loop through profiles, instead have a nodeDB function that goes directly to the map
		profiles, _ := nodeDB.FindAllProfiles()
		for _, profile := range profiles {
			wwlog.Debug("Looking for profile default: %s", profile.Id())
			if profile.Id() == "default" {
				wwlog.Debug("Found profile default, setting kernel version to: %s", args[0])
				profile.Kernel.Override = args[0]
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
