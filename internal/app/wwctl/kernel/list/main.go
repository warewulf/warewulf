package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/bufbuild/protoyaml-go"
	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

var kernelList = kernel.ListKernels

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		kernels, err := kernelList()
		if err != nil {
			wwlog.Error("%s", err)
			os.Exit(1)
		}

		nconfig, _ := node.New()
		nodes, _ := nconfig.FindAllNodes()
		nodemap := make(map[string]int)

		for _, n := range nodes {
			nodemap[n.Kernel.Override.Get()]++
		}

		headers := []string{"KERNEL NAME", "KERNEL VERSION", "NODES"}
		var kernelInfos []*wwapiv1.KernelInfo
		for _, k := range kernels {
			kernelInfos = append(kernelInfos, &wwapiv1.KernelInfo{
				KernelName:    k,
				KernelVersion: kernel.GetKernelVersion(k),
				Nodes:         strconv.Itoa(nodemap[k]),
			})
		}

		if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
			yamlBytes, err := protoyaml.Marshal(&wwapiv1.KernelListResponse{
				Kernels: kernelInfos,
			})
			if err != nil {
				return err
			}

			wwlog.Info(string(yamlBytes))
		} else if strings.EqualFold(strings.TrimSpace(vars.output), "json") {
			jsonBytes, err := json.Marshal(&wwapiv1.KernelListResponse{
				Kernels: kernelInfos,
			})
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
			for _, val := range kernelInfos {
				if err := csvWriter.Write([]string{val.KernelName, val.KernelVersion, val.Nodes}); err != nil {
					return err
				}
			}
		} else {
			ph := helper.NewPrintHelper(headers)
			for _, val := range kernelInfos {
				ph.Append([]string{val.KernelName, val.KernelVersion, val.Nodes})
			}
			ph.Render()
		}
		return nil
	}
}
