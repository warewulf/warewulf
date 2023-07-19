package delete

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	apiutil "github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/hpcng/warewulf/internal/pkg/util"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	cdp := &wwapiv1.ContainerDeleteParameter{
		ContainerNames: args,
	}
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Failed to open node database: %s", err)
		os.Exit(1)
	}

	profiles, err := nodeDB.MapAllProfiles()
	if err != nil {
		wwlog.Error("Could not load all profiles: %s", err)
		os.Exit(1)
	}

	if util.InSlice(args, profiles["default"].ContainerName.Get()) {
		return fmt.Errorf("can't delete container which is in the default profile")
	}
	if !SetYes {
		yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to delete container %s", args))
		if !yes {
			return
		}

	}
	return container.ContainerDelete(cdp)
}
