package tpm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/google/go-attestation/attest"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	target := args[0]
	var quote tpm.Quote

	if _, err := os.Stat(target); err == nil {
		data, err := os.ReadFile(target)
		if err != nil {
			return fmt.Errorf("reading quote file: %v", err)
		}

		if err := json.Unmarshal(data, &quote); err != nil {
			return fmt.Errorf("unmarshalling quote: %v", err)
		}
	} else {
		conf := config.Get()
		tpmPath := path.Join(conf.Paths.OverlayProvisiondir(), target, "tpm.json")

		data, err := os.ReadFile(tpmPath)
		if err != nil {
			return fmt.Errorf("reading tpm quote for node %s: %v", target, err)
		}

		if err := json.Unmarshal(data, &quote); err != nil {
			return fmt.Errorf("unmarshalling quote: %v", err)
		}
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
		if displayEvent {
			if err := displayEventLog(quote.EventLog); err != nil {
				wwlog.Warn("Failed to display event log: %v", err)
			}
		}
	}

	return nil
}

func displayEventLog(b64Log string) error {
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
	for _, event := range events {
		if len(pcrFilter) > 0 {
			found := false
			for _, p := range pcrFilter {
				if p == event.Index {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		fmt.Printf("PCR[%d] Type=%s Digest=%x Data=%s\n", event.Index, event.Type, event.Digest, tpm.FormatEventData(event))
	}
	return nil
}
