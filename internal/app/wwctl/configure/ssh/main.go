package ssh

import (
	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/configure"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	if len(keyTypes) == 0 {
		keyTypes = append(keyTypes, warewulfconf.Get().SSH.KeyTypes...)
	}
	return configure.SSH(keyTypes...)
}
