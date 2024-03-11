package delete

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var count int
	if util.InSlice(args, "default") {
		return fmt.Errorf("can't delete the `default` profile ")
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Failed to open node database: %s", err)
		os.Exit(1)
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Error("Could not load all profiles: %s", err)
		os.Exit(1)
	}
	for _, r := range args {
		for _, p := range profiles {
			if p.Id() == r {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					wwlog.Error("Could not load all nodes: %s", err)
					os.Exit(1)
				}
				for _, n := range nodes {
					for i, np := range n.Profiles {
						if np == r {
							wwlog.Verbose("Removing profile from node %s: %s", n.Id(), r)
							n.Profiles = append(n.Profiles[:i], n.Profiles[i+1:]...)
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
			if p.Id() == r {
				count++
				found = true
				err := nodeDB.DelProfile(r)
				if err != nil {
					wwlog.Error("%s", err)
				}
			}
		}

		if !found {
			wwlog.Warn("Profile not found: %s", r)
			return nil
		}
	}

	if count == 0 {
		wwlog.Warn("No profiles found")
		return nil
	}

	if SetYes {
		err := nodeDB.Persist()
		if err != nil {
			return errors.Wrap(err, "failed to persist nodedb")
		}
	} else {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Are you sure you want to delete %d profile(s)", count),
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
