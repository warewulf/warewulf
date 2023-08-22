package set

import (
	"fmt"
	"os"
	"strings"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) error {
		// run converters for different types
		for _, c := range vars.converters {
			if err := c(); err != nil {
				return err
			}
		}
		// remove the default network as the all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.nodeConf.NetDevs["UNDEF"]) {
			netDev := *vars.nodeConf.NetDevs["UNDEF"]
			vars.nodeConf.NetDevs[vars.netName] = &netDev
		}
		delete(vars.nodeConf.NetDevs, "UNDEF")
		if vars.fsName != "" {
			if !strings.HasPrefix(vars.fsName, "/dev") {
				if vars.fsName == vars.partName {
					vars.fsName = "/dev/disk/by-partlabel/" + vars.partName
				} else {
					return fmt.Errorf("filesystems need to have a underlying blockdev")
				}
			}
			fs := *vars.nodeConf.FileSystems["UNDEF"]
			vars.nodeConf.FileSystems[vars.fsName] = &fs
		}
		delete(vars.nodeConf.FileSystems, "UNDEF")
		if vars.diskName != "" && vars.partName != "" {
			prt := *vars.nodeConf.Disks["UNDEF"].Partitions["UNDEF"]
			vars.nodeConf.Disks["UNDEF"].Partitions[vars.partName] = &prt
			delete(vars.nodeConf.Disks["UNDEF"].Partitions, "UNDEF")
			dsk := *vars.nodeConf.Disks["UNDEF"]
			vars.nodeConf.Disks[vars.diskName] = &dsk
		}
		if (vars.diskName != "") != (vars.partName != "") {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.nodeConf.Disks, "UNDEF")
		buffer, err := yaml.Marshal(vars.nodeConf)
		if err != nil {
			wwlog.Error("Can't marshall nodeInfo", err)
			os.Exit(1)
		}
		wwlog.Debug("sending following values: %s", string(buffer))
		set := wwapiv1.NodeSetParameter{
			NodeConfYaml: string(buffer[:]),

			NetdevDelete:     vars.setNetDevDel,
			PartitionDelete:  vars.setPartDel,
			DiskDelete:       vars.setDiskDel,
			FilesystemDelete: vars.setFsDel,
			AllNodes:         vars.setNodeAll,
			Force:            vars.setForce,
			NodeNames:        args,
		}

		if !vars.setYes {
			var nodeCount uint
			// The checks run twice in the prompt case.
			// Avoiding putting in a blocking prompt in an API.
			_, nodeCount, err = apinode.NodeSetParameterCheck(&set, false)
			if err != nil {
				return nil
			}
			yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes(s)", nodeCount))
			if !yes {
				return nil
			}
		}
		return apinode.NodeSet(&set)
	}
}
