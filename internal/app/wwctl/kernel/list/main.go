package list

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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

		var kernelResp kernel.KernelListResponse
		for _, k := range kernels {
			kernelResp.Entries = append(kernelResp.Entries, &kernel.KernelListSimpleEntry{
				KernelName:    k,
				KernelVersion: kernel.GetKernelVersion(k),
				Nodes:         nodemap[k],
			})
		}

		if len(kernelResp.Entries) > 0 {
			if strings.EqualFold(strings.TrimSpace(vars.output), "yaml") {
				yamlBytes, err := yaml.Marshal(kernelResp)
				if err != nil {
					return err
				}

				wwlog.Info(string(yamlBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "json") {
				jsonBytes, err := json.Marshal(kernelResp)
				if err != nil {
					return err
				}

				wwlog.Info(string(jsonBytes))
			} else if strings.EqualFold(strings.TrimSpace(vars.output), "csv") {
				csvWriter := csv.NewWriter(os.Stdout)
				defer csvWriter.Flush()
				if err := csvWriter.Write(kernelResp.Entries[0].GetHeader()); err != nil {
					return err
				}
				for _, val := range kernelResp.Entries {
					if err := csvWriter.Write(val.GetValue()); err != nil {
						return err
					}
				}
			} else {
				ph := helper.NewPrintHelper(kernelResp.Entries[0].GetHeader())
				for _, val := range kernelResp.Entries {
					ph.Append(val.GetValue())
				}
				ph.Render()
			}
		}

		return nil
	}
}
