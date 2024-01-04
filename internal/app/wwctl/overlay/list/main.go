package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"github.com/bufbuild/protoyaml-go"
	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	apioverlay "github.com/hpcng/warewulf/internal/pkg/api/overlay"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

		var headers []string
		if vars.listLong {
			headers = []string{"PERM MODE", "UID", "GID", "SYSTEM-OVERLAY", "FILE PATH"}
		} else {
			headers = []string{"OVERLAY NAME", "FILES/DIRS"}
		}

		overlays, err := apioverlay.OverlayList(param)
		if err != nil {
			return err
		}

		if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
			yamlBytes, err := protoyaml.Marshal(overlays)
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
			if err := csvWriter.Write(headers); err != nil {
				return err
			}
			for _, val := range overlays.Overlays {
				values := util.GetProtoMessageValues(val)
				if err := csvWriter.Write(values); err != nil {
					return err
				}
			}
		} else {
			ph := helper.NewPrintHelper(headers)
			for _, val := range overlays.Overlays {
				values := util.GetProtoMessageValues(val)
				ph.Append(values)
			}
			ph.Render()
		}

		return nil
	}
}
