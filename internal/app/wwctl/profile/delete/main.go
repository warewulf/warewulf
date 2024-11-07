package delete

import (
	"fmt"

	"github.com/manifoldco/promptui"
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
		return fmt.Errorf("failed to open node database: %s", err)
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		return fmt.Errorf("could not load all profiles: %s", err)
	}
	for _, r := range args {
		for _, p := range profiles {
			if p.Id() == r {
				nodes, err := nodeDB.FindAllNodes()
				if err != nil {
					return fmt.Errorf("could not load all nodes: %s", err)
				}
				for _, n := range nodes {
					for i, np := range n.Profiles {
						if np == r {
							wwlog.Verbose("Removing profile from node %s: %s", n.Id(), r)
							n.Profiles = append(n.Profiles[:i], n.Profiles[i+1:]...)
							if err != nil {
								return fmt.Errorf("failed to update node: %w", err)
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
		}
	}

	if count == 0 {
		return fmt.Errorf("no profiles found")
	}

	if SetYes {
		err := nodeDB.Persist()
		if err != nil {
			return fmt.Errorf("failed to persist nodedb: %w", err)
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
				return fmt.Errorf("failed to persist nodedb: %w", err)
			}
		}
	}

	return nil
}
