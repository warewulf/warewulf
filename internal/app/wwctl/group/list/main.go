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
// table formatting; at least one group is required in that mode. When
// includeAll is set, the built-in `all` group is added to the listing.
func cobraRunE(noHeader, includeAll *bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %w", err)
		}

		var names []string
		seen := make(map[string]struct{}, len(args)+1)
		add := func(g string) {
			if g == "" {
				return
			}
			if _, ok := seen[g]; ok {
				return
			}
			seen[g] = struct{}{}
			names = append(names, g)
		}

		if len(args) > 0 {
			for _, a := range args {
				add(strings.TrimPrefix(a, "@"))
			}
		} else {
			for _, g := range nodeDB.ListAllGroups() {
				add(g)
			}
		}
		if *includeAll {
			add(node.AllGroup)
		}
		sort.Strings(names)

		if *noHeader {
			if len(names) == 0 {
				return fmt.Errorf("--noheader requires at least one group (pass GROUP arguments or --all)")
			}
			memberSeen := make(map[string]struct{})
			var members []string
			for _, name := range names {
				for _, m := range nodeDB.GroupMembers(name) {
					if _, ok := memberSeen[m]; ok {
						continue
					}
					memberSeen[m] = struct{}{}
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
