package add

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
RunE needs a function of type func(*cobraCommand,[]string) err, but
in order to avoid global variables which mess up testing a function of
the required type is returned
*/
func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// remove the UNDEF network as all network values are assigned
		// to this network
		if !node.ObjectIsEmpty(vars.nodeConf.NetDevs["UNDEF"]) {
			netDev := *vars.nodeConf.NetDevs["UNDEF"]
			vars.nodeConf.NetDevs[vars.nodeAdd.Net] = &netDev
			vars.nodeConf.NetDevs[vars.nodeAdd.Net].Tags = vars.nodeAdd.NetTagsAdd

		}
		delete(vars.nodeConf.NetDevs, "UNDEF")
		if vars.nodeAdd.FsName != "" {
			if !strings.HasPrefix(vars.nodeAdd.FsName, "/dev") {
				if vars.nodeAdd.FsName == vars.nodeAdd.PartName {
					vars.nodeAdd.FsName = "/dev/disk/by-partlabel/" + vars.nodeAdd.PartName
				} else {
					return fmt.Errorf("filesystems need to have a underlying blockdev")
				}
			}
			fs := *vars.nodeConf.FileSystems["UNDEF"]
			vars.nodeConf.FileSystems[vars.nodeAdd.FsName] = &fs
		}
		delete(vars.nodeConf.FileSystems, "UNDEF")
		if vars.nodeAdd.DiskName != "" && vars.nodeAdd.PartName != "" {
			prt := *vars.nodeConf.Disks["UNDEF"].Partitions["UNDEF"]
			vars.nodeConf.Disks["UNDEF"].Partitions[vars.nodeAdd.PartName] = &prt
			delete(vars.nodeConf.Disks["UNDEF"].Partitions, "UNDEF")
			dsk := *vars.nodeConf.Disks["UNDEF"]
			vars.nodeConf.Disks[vars.nodeAdd.DiskName] = &dsk
		}
		if (vars.nodeAdd.DiskName != "") != (vars.nodeAdd.PartName != "") {
			return fmt.Errorf("partition and disk must be specified")
		}
		delete(vars.nodeConf.Disks, "UNDEF")
		vars.nodeConf.Ipmi.Tags = vars.nodeAdd.IpmiTagsAdd
		if len(vars.nodeConf.Profiles) == 0 {
			if registry, err := node.New(); err == nil {
				if _, err := registry.GetProfile("default"); err == nil {
					vars.nodeConf.Profiles = []string{"default"}
				}
			}
		}
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to open node database: %w", err)
		}
		nodeArgs := hostlist.Expand(args)
		changed := cmd.Flags().Changed
		var ipv4, ipmiaddr net.IP
		for _, a := range nodeArgs {
			n, err := nodeDB.AddNode(a)
			if err != nil {
				return fmt.Errorf("failed to add node: %w", err)
			}
			n.UpdateFrom(&vars.nodeConf, changed)
			if !changed("profile") && len(vars.nodeConf.Profiles) > 0 {
				n.Profiles = vars.nodeConf.Profiles
			}
			wwlog.Info("Added node: %s", a)
			for _, dev := range n.NetDevs {
				if !ipv4.IsUnspecified() && ipv4 != nil {
					ipv4 = util.IncrementIPv4(ipv4, 1)
					wwlog.Verbose("Incremented IP addr to %s", ipv4)
					dev.Ipaddr = ipv4
				} else if !dev.Ipaddr.IsUnspecified() {
					ipv4 = dev.Ipaddr
				}
			}
			if n.Ipmi != nil {
				if !ipmiaddr.IsUnspecified() && ipmiaddr != nil {
					ipmiaddr = util.IncrementIPv4(ipmiaddr, 1)
					wwlog.Verbose("Incremented ipmi IP addr to %s", ipmiaddr)
					n.Ipmi.Ipaddr = ipmiaddr
				} else if !n.Ipmi.Ipaddr.IsUnspecified() {
					ipmiaddr = n.Ipmi.Ipaddr
				}
			}
		}
		if err := nodeDB.Persist(); err != nil {
			return fmt.Errorf("failed to persist new node: %w", err)
		}
		return warewulfd.DaemonReload()
	}
}
