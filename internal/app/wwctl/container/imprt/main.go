package imprt

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var name string
	uri := args[0]

	if len(args) == 2 {
		name = args[1]
	} else {
		name = path.Base(uri)
		fmt.Printf("Setting VNFS name: %s\n", name)
	}

	if container.ValidName(name) == false {
		wwlog.Printf(wwlog.ERROR, "VNFS name contains illegal characters: %s\n", name)
		os.Exit(1)
	}

	fullPath := container.SourceDir(name)

	if util.IsDir(fullPath) == true {
		if SetForce == true {
			wwlog.Printf(wwlog.WARN, "Overwriting existing VNFS\n")
			err := os.RemoveAll(fullPath)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		} else if SetUpdate == true {
			wwlog.Printf(wwlog.WARN, "Updating existing VNFS\n")
		} else {
			wwlog.Printf(wwlog.ERROR, "VNFS Name exists, specify --force, --update, or choose a different name: %s\n", name)
			os.Exit(1)
		}
	}

	err := container.PullURI(uri, name)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not pull image: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Building container: %s\n", name)
	output, err := container.Build(name, true)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not build container %s: %s\n", name, err)
		os.Exit(1)
	} else {
		fmt.Printf("%s: %s\n", name, output)
	}

	if SetDefault == true {
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
				wwlog.Printf(wwlog.DEBUG, "Found profile default, setting container name to: %s\n", name)
				profile.ContainerName.Set(name)
				nodeDB.ProfileUpdate(profile)
			}
		}
		nodeDB.Persist()
		fmt.Printf("Set default profile to container: %s\n", name)
		warewulfd.DaemonReload()

	}

	return nil
}
