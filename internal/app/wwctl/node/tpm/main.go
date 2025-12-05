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
	"gopkg.in/yaml.v3"
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
		tpmConfPath := path.Join(conf.Paths.Sysconfdir, "warewulf/tpm.conf")

		data, err := os.ReadFile(tpmConfPath)
		if err != nil {
			return fmt.Errorf("reading tpm config: %v", err)
		}

		var quotes map[string]tpm.Quote
		if err := yaml.Unmarshal(data, &quotes); err != nil {
			return fmt.Errorf("unmarshalling tpm config: %v", err)
		}

		var ok bool
		quote, ok = quotes[target]
		if !ok {
			return fmt.Errorf("node not found in TPM database or file not found: %s", target)
		}
	}

	if verified, err := quote.Verify(); !verified {
		wwlog.Error("Quote Verification Failed: %v", err)
		os.Exit(1)
	}

	wwlog.Info("Quote Verification Successful")

	if quote.EventLog != "" {
		if verified, err := quote.VerifyEventLog(); !verified {
			wwlog.Warn("Event Log Verification Failed: %v", err)
		} else {
			wwlog.Info("Event Log Verification Successful")
		}

		if err := displayEventLog(quote.EventLog); err != nil {
			wwlog.Warn("Failed to display event log: %v", err)
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

	fmt.Println("TPM Event Log (SHA256):")
	for _, event := range el.Events(attest.HashSHA256) {
		fmt.Printf("PCR[%d] Type=%x Digest=%x\n", event.Index, event.Type, event.Digest)
	}
	return nil
}
