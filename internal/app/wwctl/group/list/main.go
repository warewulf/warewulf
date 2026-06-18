package list

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

// cobraRunE lists all groups, or the filtered subset passed as args. Unknown
// groups are reported as having no members. When noHeader is set, the output
// is a single comma-separated, deduped list of node names with no header or
// table formatting; at least one group argument is required in that mode.
func cobraRunE(noHeader *bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if *noHeader && len(args) == 0 {
			return fmt.Errorf("--noheader requires at least one group argument")
		}

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %w", err)
		}

		var names []string
		if len(args) > 0 {
			seen := make(map[string]struct{}, len(args))
			for _, a := range args {
				a = strings.TrimPrefix(a, "@")
				if a == "" {
					continue
				}
				if _, ok := seen[a]; ok {
					continue
				}
				seen[a] = struct{}{}
				names = append(names, a)
			}
			sort.Strings(names)
		} else {
			names = nodeDB.ListAllGroups()
		}

		if *noHeader {
			seen := make(map[string]struct{})
			var members []string
			for _, name := range names {
				for _, m := range nodeDB.GroupMembers(name) {
					if _, ok := seen[m]; ok {
						continue
					}
					seen[m] = struct{}{}
					members = append(members, m)
				}
			}
			sort.Strings(members)
			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(members, ","))
			return nil
		}

		t := table.New(cmd.OutOrStdout())
		t.AddHeader("GROUP", "MEMBERS")
		for _, name := range names {
			members := nodeDB.GroupMembers(name)
			t.AddLine(table.Prep([]string{name, strings.Join(members, ",")})...)
		}
		t.Print()
		return nil
	}
}
