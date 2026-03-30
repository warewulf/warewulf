package check

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// File based verification
		for _, target := range args {
			var quote tpm.Quote
			data, err := os.ReadFile(target)
			if err != nil {
				return fmt.Errorf("reading quote file: %v", err)
			}

			if err := json.Unmarshal(data, &quote); err != nil {
				return fmt.Errorf("unmarshalling quote: %v", err)
			}
			wwlog.Info("File: %s", target)
			logStr, err := quote.VerifyAndDisplay(vars.pcrFilter, vars.displayEvent)
			if err != nil {
				return err
			}
			if logStr != "" {
				fmt.Print(logStr)
			}
		}
		return nil
	}
}
