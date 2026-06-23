package unset

import (
	"fmt"

	"github.com/spf13/cobra"
	wwctlunset "github.com/warewulf/warewulf/internal/app/wwctl/unset"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *wwctlunset.Vars) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Check if any fields were specified
		vars.NetnameChanged = cmd.Flags().Changed("netname")
		anyFieldSet := false
		for _, boolPtr := range vars.UnsetFields {
			if *boolPtr {
				anyFieldSet = true
				break
			}
		}
		anyFieldSet = anyFieldSet || len(vars.Tags) > 0 || len(vars.IpmiTags) > 0 || len(vars.NetTags) > 0 ||
			vars.NetnameChanged || cmd.Flags().Changed("diskname") ||
			cmd.Flags().Changed("fsname") || cmd.Flags().Changed("partname")
		if !anyFieldSet {
			return fmt.Errorf("no fields specified to unset")
		}

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to load node database: %w", err)
		}

		// Validate scoping: sub-entity fields require their parent scope flags
		if err := wwctlunset.ValidateScopeRequirements(vars); err != nil {
			return err
		}

		profileChanges := map[string][]node.Change{}
		modifiedCount := 0
		for _, profileName := range args {
			profilePtr, ok := nodeDB.NodeProfiles[profileName]
			if !ok {
				wwlog.Warn("invalid profile: %s", profileName)
				if !vars.UnsetForce {
					return fmt.Errorf("profile not found: %s", profileName)
				}
				continue
			}

			before := profilePtr.Clone()
			before.Flatten()

			if err := wwctlunset.UpdateEntity(profilePtr, vars); err != nil {
				return err
			}
			modifiedCount++

			profilePtr.Flatten()
			if before != nil {
				if ch := node.Diff(before, profilePtr); len(ch) > 0 {
					profileChanges[profileName] = ch
				}
			}
		}

		if modifiedCount == 0 {
			return fmt.Errorf("no profiles were modified")
		}

		summary := node.FormatChanges(profileChanges)
		if !vars.UnsetYes {
			if summary == "" {
				wwlog.Info("No changes to apply.")
				return nil
			}
			wwlog.Output("%s", summary)
			wwctlunset.WarnDeletions(vars)
			if !util.Confirm(fmt.Sprintf("Apply these changes to %d profile(s)?", len(profileChanges))) {
				wwlog.Info("No changes made!")
				return nil
			}
		} else {
			wwlog.Output("Applying following changes:\n%s", summary)
		}

		if err := nodeDB.Persist(); err != nil {
			return fmt.Errorf("failed to persist changes: %w", err)
		}

		if err := warewulfd.DaemonReload(); err != nil {
			wwlog.Warn("failed to reload daemon: %v", err)
		}

		wwlog.Info("Successfully unset fields on %d profile(s)", modifiedCount)
		return nil
	}
}
