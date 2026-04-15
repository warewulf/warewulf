package unset

import (
	"fmt"

	"github.com/spf13/cobra"
	wwctlunset "github.com/warewulf/warewulf/internal/app/wwctl/unset"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *wwctlunset.Vars) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Check if any fields were specified
		anyFieldSet := false
		for _, boolPtr := range vars.UnsetFields {
			if *boolPtr {
				anyFieldSet = true
				break
			}
		}
		anyFieldSet = anyFieldSet || len(vars.Tags) > 0 || len(vars.IpmiTags) > 0 || len(vars.NetTags) > 0 ||
			len(vars.NetDel) > 0 || len(vars.DiskDel) > 0 || len(vars.PartDel) > 0 || len(vars.FsDel) > 0
		if !anyFieldSet {
			return fmt.Errorf("no fields specified to unset")
		}

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to load node database: %w", err)
		}

		// Expand hostlist patterns
		args = hostlist.Expand(args)

		// Validate scoping: sub-entity fields require their parent scope flags
		if err := wwctlunset.ValidateScopeRequirements(vars); err != nil {
			return err
		}

		// Confirmation prompt
		if !vars.UnsetYes {
			count := 0
			for _, nodeName := range args {
				if _, ok := nodeDB.Nodes[nodeName]; ok {
					count++
				}
			}
			if count == 0 {
				return fmt.Errorf("no valid nodes found")
			}
			yes := util.Confirm(fmt.Sprintf("Are you sure you want to modify %d node(s)", count))
			if !yes {
				return nil
			}
		}

		modifiedCount := 0
		for _, nodeName := range args {
			nodePtr, ok := nodeDB.Nodes[nodeName]
			if !ok {
				wwlog.Warn("invalid node: %s", nodeName)
				if !vars.UnsetForce {
					return fmt.Errorf("node not found: %s", nodeName)
				}
				continue
			}

			if err := wwctlunset.UpdateEntity(nodePtr, vars); err != nil {
				return err
			}
			modifiedCount++
		}

		if modifiedCount == 0 {
			return fmt.Errorf("no nodes were modified")
		}

		if err := nodeDB.Persist(); err != nil {
			return fmt.Errorf("failed to persist changes: %w", err)
		}

		if err := warewulfd.DaemonReload(); err != nil {
			wwlog.Warn("failed to reload daemon: %v", err)
		}

		wwlog.Info("Successfully unset fields on %d node(s)", modifiedCount)
		return nil
	}
}
