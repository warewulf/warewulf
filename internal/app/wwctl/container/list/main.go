package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bufbuild/protoyaml-go"
	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	"github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

var containerList = container.ContainerList

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		containerInfo, err := containerList()
		if err != nil {
			wwlog.Error("%s", err)
			return
		}

		containerListResponse := &wwapiv1.ContainerListResponse{
			Containers: containerInfo,
		}

		headers := []string{"CONTAINER NAME", "NODES", "KERNEL VERSION", "CREATION TIME", "MODIFICATION TIME", "SIZE"}

		if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
			yamlBytes, err := protoyaml.Marshal(containerListResponse)
			if err != nil {
				return err
			}

			wwlog.Info(string(yamlBytes))
		} else if strings.EqualFold(strings.TrimSpace(vars.output), "json") {
			jsonBytes, err := json.Marshal(containerListResponse)
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
			for i := 0; i < len(containerInfo); i++ {
				createTime := time.Unix(int64(containerInfo[i].CreateDate), 0)
				modTime := time.Unix(int64(containerInfo[i].ModDate), 0)
				row := []string{
					containerInfo[i].Name,
					strconv.FormatUint(uint64(containerInfo[i].NodeCount), 10),
					containerInfo[i].KernelVersion,
					createTime.Format(time.RFC822),
					modTime.Format(time.RFC822),
					util.ByteToString(int64(containerInfo[i].Size)),
				}

				if err := csvWriter.Write(row); err != nil {
					return err
				}
			}
		} else {
			ph := helper.NewPrintHelper(headers)
			for i := 0; i < len(containerInfo); i++ {
				createTime := time.Unix(int64(containerInfo[i].CreateDate), 0)
				modTime := time.Unix(int64(containerInfo[i].ModDate), 0)
				ph.Append([]string{
					containerInfo[i].Name,
					strconv.FormatUint(uint64(containerInfo[i].NodeCount), 10),
					containerInfo[i].KernelVersion,
					createTime.Format(time.RFC822),
					modTime.Format(time.RFC822),
					util.ByteToString(int64(containerInfo[i].Size)),
				})
			}
			ph.Render()
		}
		return
	}
}
