package list

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path"
	"sort"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/table"
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

	// Tabular view
	t := table.New(cmd.OutOrStdout())
	
	if keyFlag {
		t.AddHeader("NODE", "MANUFACTURER", "QUOTE", "EVENTLOG", "GRUB", "SECRET")
	} else {
		t.AddHeader("NODE", "MANUFACTURER", "QUOTE", "EVENTLOG", "GRUB", "EKPUB (SHA256)")
	}

	conf := config.Get()
	for _, n := range filtered {
		tpmPath := path.Join(conf.Paths.OverlayProvisiondir(), n.Id(), "tpm.json")
		data, err := os.ReadFile(tpmPath)
		if err != nil {
			t.AddLine(table.Prep([]string{n.Id(), "", "", "", "", ""})...)
			continue
		}
		var quote tpm.Quote
		if err := json.Unmarshal(data, &quote); err != nil {
			t.AddLine(table.Prep([]string{n.Id(), "", "ERR", "", "", ""})...)
			continue
		}

		manufacturer := quote.GetManufacturer()
		quoteStatus := "PASS"
		if !quote.HasQuote() {
			quoteStatus = "N/A"
		} else if verified, err := quote.Verify(); !verified {
			wwlog.Debug("Quote Verification Failed for %s: %v", n.Id(), err)
			quoteStatus = "FAIL"
		}

		eventlogStatus := "N/A"
		grubStatus := "N/A"
		if quote.EventLog != "" {
			eventlogStatus = "PASS"
			if verified, err := quote.VerifyEventLog(); !verified {
				wwlog.Debug("Event Log Verification Failed for %s: %v", n.Id(), err)
				eventlogStatus = "FAIL"
			}

			grubStatus = "PASS"
			if err := quote.VerifyGrubBinary(); err != nil {
				wwlog.Debug("GRUB Binary Log Verification Failed for %s: %v", n.Id(), err)
				grubStatus = "FAIL"
			}
		}

		lastColStr := "N/A"
		if keyFlag {
			if quote.Challenge != nil && len(quote.Challenge.Secret) > 0 {
				lastColStr = hex.EncodeToString(quote.Challenge.Secret)
			}
		} else {
			if quote.EKPub != "" {
				if ekPubBytes, err := base64.StdEncoding.DecodeString(quote.EKPub); err == nil {
					hash := sha256.Sum256(ekPubBytes)
					lastColStr = hex.EncodeToString(hash[:])
				} else {
					lastColStr = "ERR"
				}
			}
		}

		t.AddLine(table.Prep([]string{n.Id(), manufacturer, quoteStatus, eventlogStatus, grubStatus, lastColStr})...)
	}
	t.Print()

	return nil
}
