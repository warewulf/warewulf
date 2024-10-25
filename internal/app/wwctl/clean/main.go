package clean

import (
	"github.com/warewulf/warewulf/internal/pkg/clean"

	"github.com/spf13/cobra"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		if err = clean.CleanOciBlobCacheDir(); err != nil {
			return err
		} else if err = clean.CleanOverlays(); err != nil {
			return err
		} else {
			return nil
		}
	}
}
