package list

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		if len(args) > 0 && strings.Contains(args[0], ",") {
			args = strings.FieldsFunc(args[0], func(r rune) bool { return r == ',' })
		}

		nodeDB, err := node.New()
		if err != nil {
			return
		}
		profiles, err := nodeDB.FindAllProfiles()
		if err != nil {
			return
		}
		profiles = node.FilterProfileListByName(profiles, args)
		sort.Slice(profiles, func(i, j int) bool {
			return profiles[i].Id() < profiles[j].Id()
		})

		if vars.showYaml || vars.showJson {
			profileMap := make(map[string]node.Profile)
			for _, profile := range profiles {
				profileMap[profile.Id()] = profile
			}
			var buf []byte
			if vars.showJson {
				buf, _ = json.MarshalIndent(profileMap, "", "  ")
			} else {
				buf, _ = yaml.Marshal(profileMap)
			}
			wwlog.Info(string(buf))
		} else if vars.showAll {
			t := table.New(cmd.OutOrStdout())
			t.AddHeader("PROFILE", "FIELD", "VALUE")
			for _, p := range profiles {
				fields := node.GetFieldList(p)
				for _, f := range fields {
					t.AddLine(table.Prep([]string{p.Id(), f.Field, f.Value})...)
				}
			}
			t.Print()
		} else {
			t := table.New(cmd.OutOrStdout())
			t.AddHeader("PROFILE NAME", "COMMENT/DESCRIPTION")
			for _, profile := range profiles {
				t.AddLine(table.Prep([]string{profile.Id(), profile.Comment})...)
			}
			t.Print()
		}
		return
	}
}
