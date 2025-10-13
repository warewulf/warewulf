package variables

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	overlayName := args[0]
	filePath := args[1]

	ov, err := overlay.Get(overlayName)
	if err != nil {
		wwlog.Error("Failed to get overlay %s: %s", overlayName, err)
		return err
	}

	vars := ov.ParseVars(filePath)
	if vars == nil {
		return fmt.Errorf("could not parse variables for %s in overlay %s", filePath, overlayName)
	}

	fmt.Println(strings.Join(vars, "\n"))

	return nil
}
