package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count int

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Failed to open node database: %s\n", err)
		os.Exit(1)
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Error("Could not load all profiles: %s\n", err)
		os.Exit(1)
	}

	for _, r := range args {
		for _, p := range profiles {
			if p.Id.Get() == r {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Error("Could not load all nodes: %s\n", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					for _, np := range n.Profiles.GetSlice() {
						if np == r {
							wwlog.Verbose("Removing profile from node %s: %s\n", n.Id.Get(), r)
							n.Profiles.SliceRemoveElement(r)
							err := nodeDB.NodeUpdate(n)
							if err != nil {
								return errors.Wrap(err, "failed to update node")
							}
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
					wwlog.Error("%s\n", err)
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
		err := nodeDB.Persist()
		if err != nil {
			return errors.Wrap(err, "failed to persist nodedb")
		}
	} else {
		q := fmt.Sprintf("Are you sure you want to delete %d profile(s)", count)

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			err := nodeDB.Persist()
			if err != nil {
				return errors.Wrap(err, "failed to persist nodedb")
			}
		}
	}

	return nil
}
