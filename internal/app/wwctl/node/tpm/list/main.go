package list

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

type NodeTpmStatus struct {
	Node         string `json:"node"`
	Manufacturer string `json:"manufacturer"`
	Quote        string `json:"quote"`
	EventLog     string `json:"eventlog"`
	Grub         string `json:"grub"`
	EKPub        string `json:"ekpub"`
	Secret       string `json:"secret"`
}

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

	var jsonResults []NodeTpmStatus

	conf := config.Get()
	for _, n := range filtered {
		tpmPath := path.Join(conf.Paths.OverlayProvisiondir(), n.Id(), "tpm.json")
		data, err := os.ReadFile(tpmPath)
		if err != nil {
			t.AddLine(table.Prep([]string{n.Id(), "", "", "", "", ""})...)
			jsonResults = append(jsonResults, NodeTpmStatus{Node: n.Id()})
			continue
		}
		var quote tpm.Quote
		if err := json.Unmarshal(data, &quote); err != nil {
			t.AddLine(table.Prep([]string{n.Id(), "", "ERR", "", "", ""})...)
			jsonResults = append(jsonResults, NodeTpmStatus{Node: n.Id(), Quote: "ERR"})
			continue
		}

		manufacturer := quote.Current.GetManufacturer()
		quoteStatus := "PASS"
		if !quote.Current.HasQuote() {
			quoteStatus = "N/A"
		} else if verified, err := quote.Current.Verify(); !verified {
			wwlog.Debug("Quote Verification Failed for %s: %v", n.Id(), err)
			quoteStatus = "FAIL"
		}

		eventlogStatus := "N/A"
		grubStatus := "N/A"
		if quote.EventLog != "" {
			eventlogStatus = "PASS"
			if verified, err := quote.VerifyEventLogData(&quote.Current); !verified {
				wwlog.Debug("Event Log Verification Failed for %s: %v", n.Id(), err)
				eventlogStatus = "FAIL"
			}

			grubStatus = "PASS"
			if err := quote.VerifyGrubBinaryData(&quote.Current); err != nil {
				wwlog.Debug("GRUB Binary Log Verification Failed for %s: %v", n.Id(), err)
				grubStatus = "FAIL"
			}
		}

		lastColStr := "N/A"
		ekPubStr := "N/A"
		secretStr := "N/A"
		if quote.Challenge != nil && len(quote.Challenge.Secret) > 0 {
			secretStr = hex.EncodeToString(quote.Challenge.Secret)
		}
		if quote.Current.EKPub != "" {
			if ekPubBytes, err := base64.StdEncoding.DecodeString(quote.Current.EKPub); err == nil {
				hash := sha256.Sum256(ekPubBytes)
				ekPubStr = hex.EncodeToString(hash[:])
			} else {
				ekPubStr = "ERR"
			}
		}

		if keyFlag {
			lastColStr = secretStr
		} else {
			lastColStr = ekPubStr
		}

		t.AddLine(table.Prep([]string{n.Id(), manufacturer, quoteStatus, eventlogStatus, grubStatus, lastColStr})...)
		jsonResults = append(jsonResults, NodeTpmStatus{
			Node:         n.Id(),
			Manufacturer: manufacturer,
			Quote:        quoteStatus,
			EventLog:     eventlogStatus,
			Grub:         grubStatus,
			EKPub:        ekPubStr,
			Secret:       secretStr,
		})

		if quote.New.HasQuote() {
			manufacturerNew := quote.New.GetManufacturer()
			quoteStatusNew := "PASS"
			if verified, err := quote.New.Verify(); !verified {
				wwlog.Debug("New Quote Verification Failed for %s: %v", n.Id(), err)
				quoteStatusNew = "FAIL"
			}

			eventlogStatusNew := "N/A"
			grubStatusNew := "N/A"
			if quote.EventLog != "" {
				eventlogStatusNew = "PASS"
				if verified, err := quote.VerifyEventLogData(&quote.New); !verified {
					wwlog.Debug("New Event Log Verification Failed for %s: %v", n.Id(), err)
					eventlogStatusNew = "FAIL"
				}

				grubStatusNew = "PASS"
				if err := quote.VerifyGrubBinaryData(&quote.New); err != nil {
					wwlog.Debug("New GRUB Binary Log Verification Failed for %s: %v", n.Id(), err)
					grubStatusNew = "FAIL"
				}
			}

			lastColStrNew := "N/A"
			ekPubStrNew := "N/A"
			if quote.New.EKPub != "" {
				if ekPubBytes, err := base64.StdEncoding.DecodeString(quote.New.EKPub); err == nil {
					hash := sha256.Sum256(ekPubBytes)
					ekPubStrNew = hex.EncodeToString(hash[:])
				} else {
					ekPubStrNew = "ERR"
				}
			}
			if keyFlag {
				lastColStrNew = "--"
			} else {
				lastColStrNew = ekPubStrNew
			}

			t.AddLine(table.Prep([]string{n.Id() + " (NEW)", manufacturerNew, quoteStatusNew, eventlogStatusNew, grubStatusNew, lastColStrNew})...)
		}
	}

	if jsonFlag {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(jsonResults); err != nil {
			return fmt.Errorf("failed to output JSON: %w", err)
		}
	} else {
		t.Print()
	}

	return nil
}
