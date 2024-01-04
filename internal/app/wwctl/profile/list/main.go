package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"github.com/bufbuild/protoyaml-go"
	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

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

		var headers []string
		if vars.showAll || vars.showFullAll {
			headers = []string{"PROFILE", "FIELD", "PROFILE", "VALUE"}
		} else {
			headers = []string{"PROFILE NAME", "COMMENT/DESCRIPTION"}
		}

		if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
			yamlBytes, err := protoyaml.Marshal(profiles)
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
			if err := csvWriter.Write(headers); err != nil {
				return err
			}
			for _, val := range profiles.Profiles {
				values := util.GetProtoMessageValues(val)
				if err := csvWriter.Write(values); err != nil {
					return err
				}
			}
		} else {
			ph := helper.NewPrintHelper(headers)
			for _, val := range profiles.Profiles {
				values := util.GetProtoMessageValues(val)
				ph.Append(values)
			}
			ph.Render()
		}

		return nil
	}
}
