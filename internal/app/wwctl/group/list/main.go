package list

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

// CobraRunE lists all groups, or the filtered subset passed as args. Unknown
// groups are reported as having no members.
func CobraRunE() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %w", err)
		}

		var names []string
		if len(args) > 0 {
			seen := make(map[string]struct{}, len(args))
			for _, a := range args {
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
