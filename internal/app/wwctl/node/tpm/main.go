package tpm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	"slices"
	"sort"

	"github.com/google/go-attestation/attest"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if vars.quoteFile != "" {
			// File based verification
			targets := args
			if len(targets) == 0 {
				targets = append(targets, vars.quoteFile)
			}
			for _, target := range targets {
				var quote tpm.Quote
				data, err := os.ReadFile(target)
				if err != nil {
					return fmt.Errorf("reading quote file: %v", err)
				}

				if err := json.Unmarshal(data, &quote); err != nil {
					return fmt.Errorf("unmarshalling quote: %v", err)
				}
				wwlog.Info("File: %s", target)
				if err := verifyAndDisplay(vars, &quote); err != nil {
					return err
				}
			}
			return nil
		}

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

		if vars.displayEvent {
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
				if err := verifyAndDisplay(vars, &quote); err != nil {
					wwlog.Error("Verifying node %s: %v", n.Id(), err)
				}
			}
			return nil
		}

		// Tabular view
		t := table.New(cmd.OutOrStdout())
		t.AddHeader("NODE", "MANUFACTURER", "QUOTE", "EVENTLOG", "GRUB")
		conf := config.Get()
		for _, n := range filtered {
			tpmPath := path.Join(conf.Paths.OverlayProvisiondir(), n.Id(), "tpm.json")
			data, err := os.ReadFile(tpmPath)
			if err != nil {
				t.AddLine(table.Prep([]string{n.Id(), "", "", "", ""})...)
				continue
			}
			var quote tpm.Quote
			if err := json.Unmarshal(data, &quote); err != nil {
				t.AddLine(table.Prep([]string{n.Id(), "", "ERR", "", ""})...)
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

			t.AddLine(table.Prep([]string{n.Id(), manufacturer, quoteStatus, eventlogStatus, grubStatus})...)
		}
		t.Print()

		return nil
	}
}

func verifyAndDisplay(vars *variables, quote *tpm.Quote) error {
	wwlog.Info("TPM Manufacturer: %s", quote.GetManufacturer())

	if !quote.HasQuote() {
		return fmt.Errorf("TPM Quote not available")
	}

	if verified, err := quote.Verify(); !verified {
		return fmt.Errorf("Quote Verification Failed: %v", err)
	}

	wwlog.Info("Quote Verification Successful")

	if quote.EventLog != "" {
		if verified, err := quote.VerifyEventLog(); !verified {
			return fmt.Errorf("Event Log Verification Failed: %v", err)
		} else {
			wwlog.Info("Event Log Verification Successful")
		}

		if err := quote.VerifyGrubBinary(); err != nil {
			return fmt.Errorf("GRUB Binary Log Verification Failed: %v", err)
		} else {
			wwlog.Info("GRUB Binary Log Verification Successful")
		}
		if vars.displayEvent {
			if err := displayEventLog(vars, quote.EventLog); err != nil {
				wwlog.Warn("Failed to display event log: %v", err)
			}
		}
	}
	return nil
}

func displayEventLog(vars *variables, b64Log string) error {
	logBytes, err := base64.StdEncoding.DecodeString(b64Log)
	if err != nil {
		return fmt.Errorf("decoding event log: %v", err)
	}

	el, err := attest.ParseEventLog(logBytes)
	if err != nil {
		return fmt.Errorf("parsing event log: %v", err)
	}

	events := el.Events(attest.HashSHA256)

	fmt.Println("TPM Event Log (SHA256):")
	pcrEvents := make(map[int][]string)
	for _, event := range events {
		pcrEvents[event.Index] = append(pcrEvents[event.Index], fmt.Sprintf("Type=%s Digest=%x Data=%s\n", event.Type, event.Digest, tpm.FormatEventData(event)))
	}
	for idx := range slices.Sorted(maps.Keys(pcrEvents)) {
		if len(vars.pcrFilter) > 0 {
			found := false
			for _, p := range vars.pcrFilter {
				if p == idx {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		for _, ev := range pcrEvents[idx] {
			fmt.Printf("PCR[%d] %s", idx, ev)
		}
	}
	return nil
}
