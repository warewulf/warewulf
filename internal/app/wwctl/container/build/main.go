package build

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var containers []string

	if BuildAll == true {
		containers, _ = container.ListSources()
	} else {
		containers = args
	}

	for _, c := range containers {
		if container.ValidSource(c) == false {
			wwlog.Printf(wwlog.ERROR, "VNFS name does not exist: %s\n", c)
			os.Exit(1)
		}

		container.Build(c, BuildForce)
	}

	if SetDefault == true {
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
					nodeDB.ProfileUpdate(profile)
				}
			}
			nodeDB.Persist()
			fmt.Printf("Set default profile to container: %s\n", containers[0])
		}
	}

	/*
		var nodes []node.NodeInfo
		set := make(map[string]int)

		n, err := node.New()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
			os.Exit(1)
		}

		if len(args) == 1 && ByNode == true {
			var err error
			nodes, err = n.SearchByName(args[0])
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Could not find nodes for search term: %s\n", args[0])
				os.Exit(1)
			}

			for _, node := range nodes {
				if node.Vnfs.Defined() == true {
					set[node.Vnfs.Get()]++
				}
			}

		} else if BuildAll == true {
			var err error
			nodes, err = n.FindAllNodes()
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Could not get list of nodes: %s\n", err)
				os.Exit(1)
			}

			for _, node := range nodes {
				if node.Vnfs.Defined() == true {
					wwlog.Printf(wwlog.VERBOSE, "Adding VNFS to list: %s (%s)\n", node.Vnfs.Get(), node.Id.Get())
					set[node.Vnfs.Get()]++
				}
			}

		} else if len(args) == 1 {
			set[args[0]]++
		} else {
			cmd.Usage()
			os.Exit(1)
		}

		for v := range set {
			fmt.Printf("Building VNFS: %s\n", v)
			err := container.Build(v, BuildForce)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

	*/

	return nil
}
