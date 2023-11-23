package add

import (
	"fmt"
	"os"
	"strings"

	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/spf13/cobra"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		// run converters for different types
		for _, c := range vars.Converters {
			if err := c(); err != nil {
				return err
			}
		}
		// remove the default network as the all network values are assigned
		// to this network
		if vars.netName != "" {
			netDev := *vars.profileConf.NetDevs["UNDEF"]
			vars.profileConf.NetDevs[vars.netName] = &netDev
			delete(vars.profileConf.NetDevs, "UNDEF")
		}
		// remove the UNDEF network as all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.profileConf.NetDevs["UNDEF"]) {
			netDev := *vars.profileConf.NetDevs["UNDEF"]
			vars.profileConf.NetDevs[vars.netName] = &netDev
		}
		delete(vars.profileConf.NetDevs, "UNDEF")
		if vars.fsName != "" {
			if !strings.HasPrefix(vars.fsName, "/dev") {
				if vars.fsName == vars.partName {
					vars.fsName = "/dev/disk/by-partlabel/" + vars.partName
				} else {
					return fmt.Errorf("filesystems need to have a underlying blockdev")
				}
			}
			fs := *vars.profileConf.FileSystems["UNDEF"]
			vars.profileConf.FileSystems[vars.fsName] = &fs
		}
		delete(vars.profileConf.FileSystems, "UNDEF")
		if vars.diskName != "" && vars.partName != "" {
			prt := *vars.profileConf.Disks["UNDEF"].Partitions["UNDEF"]
			vars.profileConf.Disks["UNDEF"].Partitions[vars.partName] = &prt
			delete(vars.profileConf.Disks["UNDEF"].Partitions, "UNDEF")
			dsk := *vars.profileConf.Disks["UNDEF"]
			vars.profileConf.Disks[vars.diskName] = &dsk
		}
		if (vars.diskName != "") != (vars.partName != "") {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.profileConf.Disks, "UNDEF")

		buffer, err := yaml.Marshal(vars.profileConf)
		if err != nil {
			wwlog.Error("Cant marshall nodeInfo", err)
			os.Exit(1)
		}
		set := wwapiv1.ProfileSetParameter{
			NodeConfYaml: string(buffer[:]),
			NetdevDelete: vars.SetNetDevDel,
			AllProfiles:  vars.SetNodeAll,
			Force:        vars.SetForce,
			ProfileNames: args,
		}

		if !vars.SetYes {
			// The checks run twice in the prompt case.
			// Avoiding putting in a blocking prompt in an API.
			_, _, err = apiprofile.ProfileSetParameterCheck(&set, false)
			if err != nil {
				return
			}

			yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you add the profile %s", args))
			if !yes {
				return
			}
		}
		return apiprofile.AddProfile(&set)
	}
}
