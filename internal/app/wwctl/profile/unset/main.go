package unset

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Check if any fields were specified
		anyFieldSet := false
		for _, boolPtr := range vars.unsetFields {
			if *boolPtr {
				anyFieldSet = true
				break
			}
		}
		if !anyFieldSet {
			return fmt.Errorf("no fields specified to unset")
		}

		// Load registry directly (NOT using API layer)
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to load node database: %w", err)
		}

		// Determine network name
		netname := vars.profileAdd.Net
		if netname == "" {
			netname = "default"
		}

		// Confirmation prompt
		if !vars.unsetYes {
			count := 0
			for _, profileName := range args {
				if _, ok := nodeDB.NodeProfiles[profileName]; ok {
					count++
				}
			}
			if count == 0 {
				return fmt.Errorf("no valid profiles found")
			}
			yes := util.Confirm(fmt.Sprintf("Are you sure you want to modify %d profile(s)", count))
			if !yes {
				return nil
			}
		}

		// Modify profiles directly
		modifiedCount := 0
		for _, profileName := range args {
			profilePtr, ok := nodeDB.NodeProfiles[profileName]
			if !ok {
				wwlog.Warn("invalid profile: %s", profileName)
				if !vars.unsetForce {
					return fmt.Errorf("profile not found: %s", profileName)
				}
				continue
			}

			// Zero the flagged fields using optimized single-pass reflection
			node.ApplyUnsetFieldsProfile(profilePtr, vars.unsetFields, netname)

			// Clean up empty structs
			profilePtr.Flatten()

			// Remove any empty network devices left after flattening
			// Flatten() doesn't remove individual map entries, only zeros entire fields
			for netName, netDev := range profilePtr.NetDevs {
				if node.ObjectIsEmpty(netDev) {
					delete(profilePtr.NetDevs, netName)
				}
			}

			modifiedCount++
		}

		if modifiedCount == 0 {
			return fmt.Errorf("no profiles were modified")
		}

		// Save changes
		if err := nodeDB.Persist(); err != nil {
			return fmt.Errorf("failed to persist changes: %w", err)
		}

		// Reload daemon
		if err := warewulfd.DaemonReload(); err != nil {
			wwlog.Warn("failed to reload daemon: %v", err)
			// Don't fail - changes were saved
		}

		wwlog.Info("Successfully unset fields on %d profile(s)", modifiedCount)
		return nil
	}
}
