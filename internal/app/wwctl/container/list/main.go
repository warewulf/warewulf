package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	"github.com/hpcng/warewulf/internal/pkg/api/container"
	pkg_container "github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var containerList = container.ContainerList

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		containerInfo, err := containerList()
		if err != nil {
			wwlog.Error("%s", err)
			return
		}

		resp := &pkg_container.ContainerListResponse{
			Containers: make(map[string][]pkg_container.ContainerListEntry),
		}

		if len(containerInfo) > 0 {
			for _, container := range containerInfo {
				var entries []pkg_container.ContainerListEntry
				entries = append(entries, &pkg_container.ContainerListSimpleEntry{
					Nodes:            container.NodeCount,
					KernelVersion:    container.KernelVersion,
					CreationTime:     container.CreateDate,
					ModificationTime: container.ModDate,
					Size:             container.Size,
				})

				if vals, ok := resp.Containers[container.Name]; ok {
					entries = append(entries, vals...)
				}
				resp.Containers[container.Name] = entries
			}

			if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
				yamlBytes, err := yaml.Marshal(resp)
				if err != nil {
					return err
				}

				wwlog.Info(string(yamlBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "json") {
				jsonBytes, err := json.Marshal(resp)
				if err != nil {
					return err
				}

				wwlog.Info(string(jsonBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "csv") {
				csvWriter := csv.NewWriter(os.Stdout)
				defer csvWriter.Flush()

				headerWrite := false
				for key, vals := range resp.Containers {
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
				for key, vals := range resp.Containers {
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

		return
	}
}
