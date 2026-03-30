package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Node based verification
		nodeDB, err := node.New()
		if err != nil {
			return err
		}
		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			return err
		}
		nodeNames := hostlist.Expand(args)
		sort.Strings(nodeNames)
		filtered := node.FilterNodeListByName(nodes, nodeNames)

		// Detailed view for each node
		for _, n := range filtered {
			conf := config.Get()
			tpmPath := path.Join(conf.Paths.OverlayProvisiondir(), n.Id(), "tpm.json")
			data, err := os.ReadFile(tpmPath)
			if err != nil {
				wwlog.Warn("reading tpm quote for node %s: %v", n.Id(), err)
				continue
			}
			var quote tpm.Quote
			if err := json.Unmarshal(data, &quote); err != nil {
				wwlog.Warn("unmarshalling quote for node %s: %v", n.Id(), err)
				continue
			}
			wwlog.Info("Node: %s", n.Id())
			logStr, err := quote.VerifyAndDisplay(vars.pcrFilter, vars.displayEvent)
			if err != nil {
				wwlog.Error("Verifying node %s: %v", n.Id(), err)
			} else if logStr != "" {
				fmt.Print(logStr)
			}
		}
		return nil
	}
}
