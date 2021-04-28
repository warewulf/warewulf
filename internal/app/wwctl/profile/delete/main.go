package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count int

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		os.Exit(1)
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not load all profiles: %s\n", err)
		os.Exit(1)
	}

	for _, r := range args {
		for _, p := range profiles {
			if p.Id.Get() == r {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Could not load all nodes: %s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					for _, np := range n.Profiles {
						if np == r {
							wwlog.Printf(wwlog.VERBOSE, "Removing profile from node %s: %s\n", n.Id.Get(), r)
							n.Profiles = util.SliceRemoveElement(n.Profiles, r)
							nodeDB.NodeUpdate(n)
						}
					}
				}
			}
		}
	}

	for _, r := range args {
		var found bool
		for _, p := range profiles {
			if p.Id.Get() == r {
				count++
				found = true
				err := nodeDB.DelProfile(r)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
				}
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "Profile not found: %s\n", r)
		}
	}

	if count == 0 {
		fmt.Fprintf(os.Stderr, "No profiles found\n")
		os.Exit(1)
	}

	if SetYes {
		nodeDB.Persist()
	} else {
		q := fmt.Sprintf("Are you sure you want to delete %d profile(s)", count)

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			nodeDB.Persist()
		}
	}

	return nil
}
