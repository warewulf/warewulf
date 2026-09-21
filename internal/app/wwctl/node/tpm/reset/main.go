package reset

import (
	"encoding/json"
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

func CobraRunE(cmd *cobra.Command, args []string) error {
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

	conf := config.Get()
	for _, n := range filtered {
		tpmPath := path.Join(conf.Paths.OverlayProvisiondir(), n.Id(), "tpm.json")
		data, err := os.ReadFile(tpmPath)
		if err != nil {
			if os.IsNotExist(err) {
				wwlog.Info("No TPM quote found for node %s", n.Id())
			} else {
				wwlog.Warn("reading tpm quote for node %s: %v", n.Id(), err)
			}
			continue
		}
		var quote tpm.Quote
		if err := json.Unmarshal(data, &quote); err != nil {
			wwlog.Warn("unmarshalling quote for node %s: %v", n.Id(), err)
			continue
		}

		if quote.New.HasQuote() {
			quote.Current = quote.New
			quote.New = tpm.TpmData{}
			quote.Challenge = nil

			out, err := json.MarshalIndent(quote, "", "  ")
			if err != nil {
				wwlog.Error("marshalling updated quote for node %s: %v", n.Id(), err)
				continue
			}
			if err := os.WriteFile(tpmPath, out, 0644); err != nil {
				wwlog.Error("writing updated quote for node %s: %v", n.Id(), err)
				continue
			}
			wwlog.Info("Moved NEW TPM quote to Current for node %s", n.Id())
		} else if quote.Current.HasQuote() {
			if err := os.Remove(tpmPath); err != nil && !os.IsNotExist(err) {
				wwlog.Error("removing tpm quote for node %s: %v", n.Id(), err)
				continue
			}
			wwlog.Info("Removed Current TPM quote for node %s", n.Id())
		} else {
			wwlog.Info("No active TPM quote to reset for node %s", n.Id())
		}
	}
	return nil
}
