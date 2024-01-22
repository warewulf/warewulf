package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apioverlay "github.com/hpcng/warewulf/internal/pkg/api/overlay"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		param := &wwapiv1.OverlayListParameter{}

		if len(args) > 0 {
			param.Overlays = args
		} else {
			var err error
			overlays, err := overlay.FindOverlays()
			if err != nil {
				return errors.Wrap(err, "could not obtain list of overlays from system")
			}
			param.Overlays = overlays
		}

		if vars.listLong {
			param.Type = wwapiv1.OverlayListParameter_TYPE_LONG
		} else if vars.listContents {
			param.Type = wwapiv1.OverlayListParameter_TYPE_CONTENT
		}

		overlays, err := apioverlay.OverlayList(param)
		if err != nil {
			return err
		}

		if len(overlays.Overlays) > 0 {
			if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
				yamlBytes, err := yaml.Marshal(overlays)
				if err != nil {
					return err
				}

				wwlog.Info(string(yamlBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "json") {
				jsonBytes, err := json.Marshal(overlays)
				if err != nil {
					return err
				}

				wwlog.Info(string(jsonBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "csv") {

				csvWriter := csv.NewWriter(os.Stdout)
				defer csvWriter.Flush()

				headerWrite := false
				for key, vals := range overlays.Overlays {
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
				for key, vals := range overlays.Overlays {
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
