//go:build !linux
// +build !linux

package child

import (
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	wwlog.Error("This command does not work on non-Linux hosts")

	return nil
}
