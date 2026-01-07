package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// ValidateUnsetScope checks that sub-entity unset flags have the required
// scoping flags. Disk field flags require --diskname; partition field flags
// require both --diskname and --partname; filesystem field flags require --fsname.
func ValidateUnsetScope(unsetFields map[string]*bool, diskName, partName, fsName string) error {
	for flagName, boolPtr := range unsetFields {
		if boolPtr == nil || !*boolPtr {
			continue
		}
		if strings.HasPrefix(flagName, "part") {
			if diskName == "" || partName == "" {
				return fmt.Errorf("--diskname and --partname must be specified with --%s", flagName)
			}
		} else if strings.HasPrefix(flagName, "disk") {
			if diskName == "" {
				return fmt.Errorf("--diskname must be specified with --%s", flagName)
			}
		} else if strings.HasPrefix(flagName, "fs") {
			if fsName == "" {
				return fmt.Errorf("--fsname must be specified with --%s", flagName)
			}
		}
	}
	return nil
}

func AddContainer(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "container", "C", "", "Set image name (backwards-compatibility)")
	cmd.Flags().Lookup("container").Hidden = true
	if err := cmd.Flags().MarkDeprecated("container", "use --image instead"); err != nil {
		panic(err)
	}
}

func AddWwinit(cmd *cobra.Command, dest *[]string) {
	cmd.Flags().StringSliceVar(dest, "wwinit", []string{}, "Set the system overlay")
	cmd.Flags().Lookup("wwinit").Hidden = true
	if err := cmd.Flags().MarkDeprecated("wwinit", "use --system-overlays instead"); err != nil {
		panic(err)
	}
}

func AddRuntime(cmd *cobra.Command, dest *[]string) {
	cmd.Flags().StringSliceVar(dest, "runtime", []string{}, "Set the runtime overlay")
	cmd.Flags().Lookup("runtime").Hidden = true
	if err := cmd.Flags().MarkDeprecated("runtime", "use --runtime-overlays instead"); err != nil {
		panic(err)
	}
}
