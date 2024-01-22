package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/spf13/cobra"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		if len(args) > 0 && strings.Contains(args[0], ",") {
			args = strings.FieldsFunc(args[0], func(r rune) bool { return r == ',' })
		}
		req := wwapiv1.GetProfileList{
			ShowAll:     vars.showAll,
			ShowFullAll: vars.showFullAll,
			Profiles:    args,
		}

		profiles, err := apiprofile.ProfileList(&req)
		if err != nil {
			return
		}

		if len(profiles.Profiles) > 0 {
			if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
				yamlBytes, err := yaml.Marshal(profiles)
				if err != nil {
					return err
				}

				wwlog.Info(string(yamlBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "json") {
				jsonBytes, err := json.Marshal(profiles)
				if err != nil {
					return err
				}

				wwlog.Info(string(jsonBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "csv") {
				csvWriter := csv.NewWriter(os.Stdout)
				defer csvWriter.Flush()

				headerWrite := false

				// sort the keys for output
				keys := make([]string, 0, len(profiles.Profiles))
				for k := range profiles.Profiles {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, key := range keys {
					vals := profiles.Profiles[key]
					if !headerWrite {
						if err := csvWriter.Write(vals[0].GetHeader()); err != nil {
							return err
						}
						headerWrite = true
					}

					for _, val := range vals {
						columns := []string{key}
						columns = append(columns, val.GetValue()...)
						if err := csvWriter.Write(columns); err != nil {
							return err
						}
					}
				}
			} else {
				var ph *helper.PrintHelper
				headerWrite := false

				// sort the keys for output
				keys := make([]string, 0, len(profiles.Profiles))
				for k := range profiles.Profiles {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, key := range keys {
					vals := profiles.Profiles[key]
					if !headerWrite {
						ph = helper.NewPrintHelper(vals[0].GetHeader())
					}
					headerWrite = true

					for _, val := range vals {
						columns := []string{key}
						columns = append(columns, val.GetValue()...)
						ph.Append(columns)
					}
				}
				ph.Render()
			}
		}

		return nil
	}
}
