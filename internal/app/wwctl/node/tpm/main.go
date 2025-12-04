package tpm

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	quoteFile := args[0]

	data, err := os.ReadFile(quoteFile)
	if err != nil {
		return fmt.Errorf("reading quote file: %v", err)
	}

	var quote tpm.Quote
	if err := json.Unmarshal(data, &quote); err != nil {
		return fmt.Errorf("unmarshalling quote: %v", err)
	}

	if err := tpm.VerifyQuote(&quote); err != nil {
		wwlog.Error("Quote Verification Failed: %v", err)
		os.Exit(1)
	}

	wwlog.Info("Quote Verification Successful")
	return nil
}
