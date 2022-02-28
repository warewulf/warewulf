package syncuser

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	containerName := args[0]
	if !container.ValidName(containerName) {
		return fmt.Errorf("%s is not a valid container", containerName)
	}
	err := container.SyncUids(containerName)
	if err != nil {
		fmt.Sprint(err)
		os.Exit(1)
	}

	return nil
}
