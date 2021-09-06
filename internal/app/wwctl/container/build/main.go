package build

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var containers []string

	if BuildAll {
		containers, _ = container.ListSources()
	} else {
		containers = args
	}

	if len(containers) == 0 {
		fmt.Println(cmd.Help())
		os.Exit(0)
	}

	for _, c := range containers {
		if !container.ValidSource(c) {
			wwlog.Printf(wwlog.ERROR, "VNFS name does not exist: %s\n", c)
			os.Exit(1)
		}

		err := container.Build(c, BuildForce)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not build container %s: %s\n", c, err)
			os.Exit(1)
		}
	}

	if SetDefault {
		if len(containers) != 1 {
			wwlog.Printf(wwlog.ERROR, "Can only set default for one container\n")
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
					wwlog.Printf(wwlog.DEBUG, "Found profile default, setting container name to: %s\n", containers[0])
					profile.ContainerName.Set(containers[0])
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
			fmt.Printf("Set default profile to container: %s\n", containers[0])
		}
	}

	return nil
}
