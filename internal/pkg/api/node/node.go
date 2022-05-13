package node

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
)

// NodeAdd adds nodes for management by Warewulf.
func NodeAdd(nap *wwapiv1.NodeAddParameter) (err error) {

	if nap == nil {
		return fmt.Errorf("NodeAddParameter is nil")
	}

	var count uint
	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "failed to open node database")
	}

	node_args := hostlist.Expand(nap.NodeNames)

	for _, a := range node_args {
		var n node.NodeInfo
		n, err = nodeDB.AddNode(a)
		if err != nil {
			return errors.Wrap(err, "failed to add node")
		}
		wwlog.Printf(wwlog.INFO, "Added node: %s\n", a)

		if nap.Cluster != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting cluster name to: %s\n", n.Id.Get(), nap.Cluster)
			n.ClusterName.Set(nap.Cluster)
			err = nodeDB.NodeUpdate(n)
			if err != nil {
				return errors.Wrap(err, "failed to update node")
			}
		}

		if nap.Netdev != "" {
			err = checkNetNameRequired(nap.Netname)
			if err != nil {
				return
			}

			if _, ok := n.NetDevs[nap.Netname]; !ok {
				var netdev node.NetDevEntry
				n.NetDevs[nap.Netname] = &netdev
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Device to: %s\n", n.Id.Get(), nap.Netname, nap.Netdev)

			n.NetDevs[nap.Netname].Device.Set(nap.Netdev)
			n.NetDevs[nap.Netname].OnBoot.SetB(true)
		}

		if nap.Ipaddr != "" {
			err = checkNetNameRequired(nap.Netname)
			if err != nil {
				return
			}

			NewIpaddr := util.IncrementIPv4(nap.Ipaddr, count)

			if _, ok := n.NetDevs[nap.Netname]; !ok {
				var netdev node.NetDevEntry
				n.NetDevs[nap.Netname] = &netdev
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Id.Get(), nap.Netname, NewIpaddr)

			n.NetDevs[nap.Netname].Ipaddr.Set(NewIpaddr)
			n.NetDevs[nap.Netname].OnBoot.SetB(true)
		}

		if nap.Netmask != "" {
			err = checkNetNameRequired(nap.Netname)
			if err != nil {
				return
			}

			if _, ok := n.NetDevs[nap.Netname]; !ok {
				return errors.New("network device does not exist: " + nap.Netname)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Id.Get(), nap.Netname, nap.Netmask)

			n.NetDevs[nap.Netname].Netmask.Set(nap.Netmask)
		}

		if nap.Gateway != "" {
			err = checkNetNameRequired(nap.Netname)
			if err != nil {
				return
			}

			if _, ok := n.NetDevs[nap.Netname]; !ok {
				return errors.New("network device does not exist: " + nap.Netname)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Id.Get(), nap.Netname, nap.Gateway)

			n.NetDevs[nap.Netname].Gateway.Set(nap.Gateway)
		}

		if nap.Hwaddr != "" {
			if nap.Netname == "" {
				return errors.New("you must include the '--netname' option")
			}

			if _, ok := n.NetDevs[nap.Netname]; !ok {
				return errors.New("network device does not exist: " + nap.Netname)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Id.Get(), nap.Netname, nap.Hwaddr)

			n.NetDevs[nap.Netname].Hwaddr.Set(nap.Hwaddr)
			n.NetDevs[nap.Netname].OnBoot.SetB(true)
		}

		if nap.Type != "" {
			if nap.Netname == "" {
				return errors.New("you must include the '--netname' option")
			}

			if _, ok := n.NetDevs[nap.Netname]; !ok {
				return errors.New("network device does not exist: " + nap.Netname)
			}
			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Type to: %s\n", n.Id.Get(), nap.Netname, nap.Type)

			n.NetDevs[nap.Netname].Type.Set(nap.Type)
		}

		if nap.Ipaddr6 != "" {
			if nap.Netname == "" {
				return errors.New("you must include the '--netname' option")
			}
			if _, ok := n.NetDevs[nap.Netname]; !ok {
				return errors.New("network device does not exist: " + nap.Netname)
			}
			// just check if address is a valid ipv6 CIDR address
			if _, _, err := net.ParseCIDR(nap.Ipaddr6); err != nil {
				return errors.New(fmt.Sprintf("%s is not a valid ipv6 address in CIDR notation\n", nap.Ipaddr6))
			}
		}

		if nap.Discoverable {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to discoverable\n", n.Id.Get())

			n.Discoverable.SetB(true)
		}

		err = nodeDB.NodeUpdate(n)
		if err != nil {
			return errors.Wrap(err, "failed to update nodedb")
		}

		count++
	} // end for

	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist new node")
	}

	err = warewulfd.DaemonReload()
	if err != nil {
		return errors.Wrap(err, "failed to reload warewulf daemon")
	}
	return
}

// NodeDelete adds nodes for management by Warewulf.
func NodeDelete(ndp *wwapiv1.NodeDeleteParameter) (err error) {

	var nodeList []node.NodeInfo
	nodeList, err = NodeDeleteParameterCheck(ndp, false)
	if err != nil {
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		return
	}

	for _, n := range nodeList {
		err := nodeDB.DelNode(n.Id.Get())
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
		} else {
			//count++
			fmt.Printf("Deleting node: %s\n", n.Id.Print())
		}
	}

	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist nodedb")
	}

	err = warewulfd.DaemonReload()
	if err != nil {
		return errors.Wrap(err, "failed to reload warewulf daemon")
	}
	return
}

// NodeDeleteParameterCheck does error checking on NodeDeleteParameter.
// Output to the console if console is true.
// Returns the nodes to delete.
func NodeDeleteParameterCheck(ndp *wwapiv1.NodeDeleteParameter, console bool) (nodeList []node.NodeInfo, err error) {

	if ndp == nil {
		err = fmt.Errorf("NodeDeleteParameter is nil")
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed to open node database: %s\n", err)
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
		return
	}

	node_args := hostlist.Expand(ndp.NodeNames)

	for _, r := range node_args {
		var match bool
		for _, n := range nodes {
			if n.Id.Get() == r {
				nodeList = append(nodeList, n)
				match = true
			}
		}

		if !match {
			fmt.Fprintf(os.Stderr, "ERROR: No match for node: %s\n", r)
		}
	}

	if len(nodeList) == 0 {
		fmt.Printf("No nodes found\n")
	}
	return
}

// NodeList lists all to none of the nodes managed by Warewulf.
func NodeList(nodeNames []string) (nodeInfo []*wwapiv1.NodeInfo, err error) {

	// nil is okay for nodeNames

	nodeDB, err := node.New()
	if err != nil {
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return
	}

	nodeNames = hostlist.Expand(nodeNames)

	// Translate to the protobuf structure so wwapiv1 can use this across the wire.
	// This is the same logic as was in wwctl.
	for _, node := range node.FilterByName(nodes, nodeNames) {

		var ni wwapiv1.NodeInfo

		ni.Id = &wwapiv1.NodeField{
			Source: node.Id.Source(),
			Value:  node.Id.Get(),
			Print:  node.Id.Print(),
		}

		ni.Comment = &wwapiv1.NodeField{
			Source: node.Comment.Source(),
			Value:  node.Comment.Get(),
			Print:  node.Comment.Print(),
		}

		ni.Cluster = &wwapiv1.NodeField{
			Source: node.ClusterName.Source(),
			Value:  node.ClusterName.Get(),
			Print:  node.ClusterName.Print(),
		}

		ni.Profiles = node.Profiles

		ni.Discoverable = &wwapiv1.NodeField{
			Source: node.Discoverable.Source(),
			Value:  strconv.FormatBool(node.Discoverable.GetB()),
			Print:  node.Discoverable.PrintB(),
		}

		ni.Container = &wwapiv1.NodeField{
			Source: node.ContainerName.Source(),
			Value:  node.ContainerName.Get(),
			Print:  node.ContainerName.Print(),
		}

		ni.KernelOverride = &wwapiv1.NodeField{
			Source: node.Kernel.Override.Source(),
			Value:  node.Kernel.Override.Get(),
			Print:  node.Kernel.Override.Print(),
		}

		ni.KernelArgs = &wwapiv1.NodeField{
			Source: node.Kernel.Args.Source(),
			Value:  node.Kernel.Args.Get(),
			Print:  node.Kernel.Args.Print(),
		}

		ni.SystemOverlay = &wwapiv1.NodeField{
			Source: node.SystemOverlay.Source(),
			Value:  node.SystemOverlay.Get(),
			Print:  node.SystemOverlay.Print(),
		}

		ni.RuntimeOverlay = &wwapiv1.NodeField{
			Source: node.RuntimeOverlay.Source(),
			Value:  node.RuntimeOverlay.Get(),
			Print:  node.RuntimeOverlay.Print(),
		}

		ni.Ipxe = &wwapiv1.NodeField{
			Source: node.Ipxe.Source(),
			Value:  node.Ipxe.Get(),
			Print:  node.Ipxe.Print(),
		}

		ni.Init = &wwapiv1.NodeField{
			Source: node.Init.Source(),
			Value:  node.Init.Get(),
			Print:  node.Init.Print(),
		}

		ni.Root = &wwapiv1.NodeField{
			Source: node.Root.Source(),
			Value:  node.Root.Get(),
			Print:  node.Root.Print(),
		}

		ni.AssetKey = &wwapiv1.NodeField{
			Source: node.AssetKey.Source(),
			Value:  node.AssetKey.Get(),
			Print:  node.AssetKey.Print(),
		}

		ni.IpmiIpaddr = &wwapiv1.NodeField{
			Source: node.Ipmi.Ipaddr.Source(),
			Value:  node.Ipmi.Ipaddr.Get(),
			Print:  node.Ipmi.Ipaddr.Print(),
		}

		ni.IpmiNetmask = &wwapiv1.NodeField{
			Source: node.Ipmi.Netmask.Source(),
			Value:  node.Ipmi.Netmask.Get(),
			Print:  node.Ipmi.Netmask.Print(),
		}

		ni.IpmiPort = &wwapiv1.NodeField{
			Source: node.Ipmi.Port.Source(),
			Value:  node.Ipmi.Port.Get(),
			Print:  node.Ipmi.Port.Print(),
		}

		ni.IpmiGateway = &wwapiv1.NodeField{
			Source: node.Ipmi.Gateway.Source(),
			Value:  node.Ipmi.Gateway.Get(),
			Print:  node.Ipmi.Gateway.Print(),
		}

		ni.IpmiUserName = &wwapiv1.NodeField{
			Source: node.Ipmi.UserName.Source(),
			Value:  node.Ipmi.UserName.Get(),
			Print:  node.Ipmi.UserName.Print(),
		}

		ni.IpmiPassword = &wwapiv1.NodeField{
			Source: node.Ipmi.Password.Source(),
			Value:  node.Ipmi.Password.Get(),
			Print:  node.Ipmi.Password.Print(), // TODO: Password was removed from pprinted output, at least in some places.
		}

		ni.IpmiInterface = &wwapiv1.NodeField{
			Source: node.Ipmi.Interface.Source(),
			Value:  node.Ipmi.Interface.Get(),
			Print:  node.Ipmi.Interface.Print(),
		}

		for keyname, keyvalue := range node.Tags {
			ni.Tags[keyname].Source = keyvalue.Source()
			ni.Tags[keyname].Value = keyvalue.Get()
			ni.Tags[keyname].Print = keyvalue.Print()
		}

		ni.NetDevs = map[string]*wwapiv1.NetDev{}
		for name, netdev := range node.NetDevs {

			ni.NetDevs[name] = &wwapiv1.NetDev{
				Device: &wwapiv1.NodeField{
					Source: netdev.Device.Source(),
					Value:  netdev.Device.Get(),
					Print:  netdev.Device.Print(),
				},
				Hwaddr: &wwapiv1.NodeField{
					Source: netdev.Hwaddr.Source(),
					Value:  netdev.Hwaddr.Get(),
					Print:  netdev.Hwaddr.Print(),
				},
				Ipaddr: &wwapiv1.NodeField{
					Source: netdev.Ipaddr.Source(),
					Value:  netdev.Ipaddr.Get(),
					Print:  netdev.Ipaddr.Print(),
				},
				Netmask: &wwapiv1.NodeField{
					Source: netdev.Netmask.Source(),
					Value:  netdev.Netmask.Get(),
					Print:  netdev.Netmask.Print(),
				},
				Gateway: &wwapiv1.NodeField{
					Source: netdev.Gateway.Source(),
					Value:  netdev.Gateway.Get(),
					Print:  netdev.Gateway.Print(),
				},
				Type: &wwapiv1.NodeField{
					Source: netdev.Type.Source(),
					Value:  netdev.Type.Get(),
					Print:  netdev.Type.Print(),
				},
				Onboot: &wwapiv1.NodeField{
					Source: netdev.OnBoot.Source(),
					Value:  strconv.FormatBool(netdev.OnBoot.GetB()),
					Print:  netdev.OnBoot.PrintB(),
				},
				Primary: &wwapiv1.NodeField{
					Source: netdev.Primary.Source(),
					Value:  strconv.FormatBool(netdev.Primary.GetB()),
					Print:  netdev.Primary.PrintB(),
				},
			}
		}
		nodeInfo = append(nodeInfo, &ni)
	}
	return
}

// NodeSet is the wwapiv1 implmentation for updating node fields.
func NodeSet(set *wwapiv1.NodeSetParameter) (err error) {

	if set == nil {
		return fmt.Errorf("NodeAddParameter is nil")
	}

	var nodeDB node.NodeYaml
	nodeDB, _, err = NodeSetParameterCheck(set, false)
	if err != nil {
		return
	}
	return nodeDbSave(&nodeDB)
}

// NodeSetParameterCheck does error checking on NodeSetParameter.
// Output to the console if console is true.
// TODO: Determine if the console switch does wwlog or not.
// - console may end up being textOutput?
func NodeSetParameterCheck(set *wwapiv1.NodeSetParameter, console bool) (nodeDB node.NodeYaml, nodeCount uint, err error) {

	if set == nil {
		err = fmt.Errorf("Node set parameter is nil")
		if console {
			fmt.Printf("%v\n", err)
			return
		}
	}

	if set.NodeNames == nil {
		err = fmt.Errorf("Node set parameter: NodeNames is nil")
		if console {
			fmt.Printf("%v\n", err)
			return
		}
	}

	var setProfiles []string
	nodeDB, err = node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
		return
	}

	// Note: This does not do expansion on the nodes.

	if set.AllNodes || (len(set.NodeNames) == 0 && len(nodes) > 0) {
		if console {
			fmt.Printf("\n*** WARNING: This command will modify all nodes! ***\n\n")
		}
	} else {
		nodes = node.FilterByName(nodes, set.NodeNames)
	}

	if len(nodes) == 0 {
		if console {
			fmt.Printf("No nodes found\n")
		}
		return
	}

	for _, n := range nodes {
		wwlog.Printf(wwlog.VERBOSE, "Evaluating node: %s\n", n.Id.Get())

		if set.Comment != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting comment to: %s\n", n.Id.Get(), set.Comment)
			n.Comment.Set(set.Comment)
		}

		if set.Container != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting container name to: %s\n", n.Id.Get(), set.Container)
			n.ContainerName.Set(set.Container)
		}

		if set.Init != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting init command to: %s\n", n.Id.Get(), set.Init)
			n.Init.Set(set.Init)
		}

		if set.Root != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting root to: %s\n", n.Id.Get(), set.Root)
			n.Root.Set(set.Root)
		}

		if set.AssetKey != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting asset key to: %s\n", n.Id.Get(), set.AssetKey)
			n.AssetKey.Set(set.AssetKey)
		}

		if set.KernelOverride != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting kernel override to: %s\n", n.Id.Get(), set.KernelOverride)
			n.Kernel.Override.Set(set.KernelOverride)
		}

		if set.KernelArgs != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting kernel args to: %s\n", n.Id.Get(), set.KernelArgs)
			n.Kernel.Args.Set(set.KernelArgs)
		}

		if set.Cluster != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting cluster name to: %s\n", n.Id.Get(), set.Cluster)
			n.ClusterName.Set(set.Cluster)
		}

		if set.Ipxe != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting iPXE template to: %s\n", n.Id.Get(), set.Ipxe)
			n.Ipxe.Set(set.Ipxe)
		}

		if len(set.RuntimeOverlay) != 0 {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting runtime overlay to: %s\n", n.Id.Get(), set.RuntimeOverlay)
			n.RuntimeOverlay.SetSlice(set.RuntimeOverlay)
		}

		if len(set.SystemOverlay) != 0 {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting system overlay to: %s\n", n.Id.Get(), set.SystemOverlay)
			n.SystemOverlay.SetSlice(set.SystemOverlay)
		}

		if set.IpmiIpaddr != "" {
			newIpaddr := util.IncrementIPv4(set.IpmiIpaddr, nodeCount)
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP address to: %s\n", n.Id.Get(), newIpaddr)
			n.Ipmi.Ipaddr.Set(newIpaddr)
		}

		if set.IpmiNetmask != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI netmask to: %s\n", n.Id.Get(), set.IpmiNetmask)
			n.Ipmi.Netmask.Set(set.IpmiNetmask)
		}

		if set.IpmiPort != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI port to: %s\n", n.Id.Get(), set.IpmiPort)
			n.Ipmi.Port.Set(set.IpmiPort)
		}

		if set.IpmiGateway != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI gateway to: %s\n", n.Id.Get(), set.IpmiGateway)
			n.Ipmi.Gateway.Set(set.IpmiGateway)
		}

		if set.IpmiUsername != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP username to: %s\n", n.Id.Get(), set.IpmiUsername)
			n.Ipmi.UserName.Set(set.IpmiUsername)
		}

		if set.IpmiPassword != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP password to: %s\n", n.Id.Get(), set.IpmiPassword)
			n.Ipmi.Password.Set(set.IpmiPassword)
		}

		if set.IpmiInterface != "" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting IPMI IP interface to: %s\n", n.Id.Get(), set.IpmiInterface)
			n.Ipmi.Interface.Set(set.IpmiInterface)
		}

		if set.IpmiWrite == "yes" || set.Onboot == "y" || set.Onboot == "1" || set.Onboot == "true" {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting Ipmiwrite to %s\n", n.Id.Get(), set.IpmiWrite)
			n.Ipmi.Write.SetB(true)
		} else {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting Ipmiwrite to %s\n", n.Id.Get(), set.IpmiWrite)
			n.Ipmi.Write.SetB(false)
		}

		if set.Discoverable {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to discoverable\n", n.Id.Get())
			n.Discoverable.SetB(true)
		}

		if set.Undiscoverable {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting node to undiscoverable\n", n.Id.Get())
			n.Discoverable.SetB(false)
		}

		if len(setProfiles) > 0 {
			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting profiles to: %s\n", n.Id.Get(), strings.Join(setProfiles, ","))
			n.Profiles = setProfiles
		}

		if len(set.ProfileAdd) > 0 {
			for _, p := range set.ProfileAdd {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, adding profile '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceAddUniqueElement(n.Profiles, p)
			}
		}

		if len(set.ProfileDelete) > 0 {
			for _, p := range set.ProfileDelete {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s, deleting profile '%s'\n", n.Id.Get(), p)
				n.Profiles = util.SliceRemoveElement(n.Profiles, p)
			}
		}

		if set.Netname != "" {
			if _, ok := n.NetDevs[set.Netname]; !ok {
				var nd node.NetDevEntry

				n.NetDevs[set.Netname] = &nd

				if set.Netdev == "" {
					n.NetDevs[set.Netname].Device.Set(set.Netname)
				}
			}
			var def bool = true

			// NOTE: This is overriding parameters passed in by the caller.
			set.Onboot = "yes"

			for _, n := range n.NetDevs {
				if n.Primary.GetB() {
					def = false
				}
			}

			if def {
				set.NetDefault = "yes"
			}
		}

		if set.Netdev != "" {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting net Device to: %s\n", n.Id.Get(), set.Netname, set.Netdev)
			n.NetDevs[set.Netname].Device.Set(set.Netdev)
		}

		if set.Ipaddr != "" {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			newIpaddr := util.IncrementIPv4(set.Ipaddr, nodeCount)

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Ipaddr to: %s\n", n.Id.Get(), set.Netname, newIpaddr)
			n.NetDevs[set.Netname].Ipaddr.Set(newIpaddr)
		}

		if set.Netmask != "" {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting netmask to: %s\n", n.Id.Get(), set.Netname, set.Netmask)
			n.NetDevs[set.Netname].Netmask.Set(set.Netmask)
		}

		if set.Gateway != "" {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting gateway to: %s\n", n.Id.Get(), set.Netname, set.Gateway)
			n.NetDevs[set.Netname].Gateway.Set(set.Gateway)
		}

		if set.Hwaddr != "" {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting HW address to: %s\n", n.Id.Get(), set.Netname, set.Hwaddr)
			n.NetDevs[set.Netname].Hwaddr.Set(strings.ToLower(set.Hwaddr))
		}

		if set.Type != "" {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting Type %s\n", n.Id.Get(), set.Netname, set.Type)
			n.NetDevs[set.Netname].Type.Set(set.Type)
		}

		if set.Onboot != "" {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			if set.Onboot == "yes" || set.Onboot == "y" || set.Onboot == "1" || set.Onboot == "true" {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting ONBOOT\n", n.Id.Get(), set.Netname)
				n.NetDevs[set.Netname].OnBoot.SetB(true)
			} else {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Unsetting ONBOOT\n", n.Id.Get(), set.Netname)
				n.NetDevs[set.Netname].OnBoot.SetB(false)
			}
		}

		if set.NetDefault != "" {
			if set.Netname == "" {
				err = fmt.Errorf("You must include the '--netname' option")
				wwlog.Printf(wwlog.ERROR, fmt.Sprintf("%v\n", err.Error()))
				return
			}

			if set.NetDefault == "yes" || set.NetDefault == "y" || set.NetDefault == "1" || set.NetDefault == "true" {

				// Set all other devices to non-default
				for _, n := range n.NetDevs {
					n.Primary.SetB(false)
				}

				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Setting PRIMARY\n", n.Id.Get(), set.Netname)
				n.NetDevs[set.Netname].Primary.SetB(true)
			} else {
				wwlog.Printf(wwlog.VERBOSE, "Node: %s:%s, Unsetting PRIMARY\n", n.Id.Get(), set.Netname)
				n.NetDevs[set.Netname].Primary.SetB(false)
			}
		}

		if set.NetdevDelete {
			err = checkNetNameRequired(set.Netname)
			if err != nil {
				return
			}

			if _, ok := n.NetDevs[set.Netname]; !ok {
				err = fmt.Errorf("Network device name doesn't exist: %s", set.Netname)
				wwlog.Printf(wwlog.ERROR, fmt.Sprintf("%v\n", err.Error()))
				return
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting network device: %s\n", n.Id.Get(), set.Netname)
			delete(n.NetDevs, set.Netname)
		}

		if len(set.Tags) > 0 {
			for _, t := range set.Tags {
				keyval := strings.SplitN(t, "=", 2)
				key := keyval[0]
				val := keyval[1]

				if _, ok := n.Tags[key]; !ok {
					var nd node.Entry
					n.Tags[key] = &nd
				}

				wwlog.Printf(wwlog.VERBOSE, "Node: %s, Setting Tag '%s'='%s'\n", n.Id.Get(), key, val)
				n.Tags[key].Set(val)
			}
		}
		if len(set.TagsDelete) > 0 {
			for _, t := range set.TagsDelete {
				keyval := strings.SplitN(t, "=", 1)
				key := keyval[0]

				if _, ok := n.Tags[key]; !ok {
					wwlog.Printf(wwlog.WARN, "Key does not exist: %s\n", key)
					os.Exit(1)
				}

				wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting tag: %s\n", n.Id.Get(), key)
				delete(n.Tags, key)
			}
		}

		err := nodeDB.NodeUpdate(n)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		nodeCount++
	}
	return
}

// NodeStatus returns the imaging state for nodes.
// This requires warewulfd.
func NodeStatus(nodeNames []string) (nodeStatusResponse *wwapiv1.NodeStatusResponse, err error) {

	// Local structs for translating json from warewulfd.
	type nodeStatusInternal struct {
		NodeName string `json:"node name"`
		Stage    string `json:"stage"`
		Sent     string `json:"sent"`
		Ipaddr   string `json:"ipaddr"`
		Lastseen int64  `json:"last seen"`
	}

	// all status is a map with one key (nodes)
	// and maps of [nodeName]NodeStatus underneath.
	type allStatus struct {
		Nodes map[string]*nodeStatusInternal `json:"nodes"`
	}

	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		return
	}

	if controller.Ipaddr == "" {
		err = fmt.Errorf("The Warewulf Server IP Address is not properly configured")
		wwlog.Printf(wwlog.ERROR, fmt.Sprintf("%v\n", err.Error()))
		return
	}

	statusURL := fmt.Sprintf("http://%s:%d/status", controller.Ipaddr, controller.Warewulf.Port)
	wwlog.Printf(wwlog.VERBOSE, "Connecting to: %s\n", statusURL)

	resp, err := http.Get(statusURL)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not connect to Warewulf server: %s\n", err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var wwNodeStatus allStatus

	err = decoder.Decode(&wwNodeStatus)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not decode JSON: %s\n", err)
		return
	}

	// Translate struct and filter.
	nodeStatusResponse = &wwapiv1.NodeStatusResponse{}

	if len(nodeNames) == 0 {
		for _, v := range wwNodeStatus.Nodes {
			nodeStatusResponse.NodeStatus = append(nodeStatusResponse.NodeStatus,
				&wwapiv1.NodeStatus{
					NodeName: v.NodeName,
					Stage:    v.Stage,
					Sent:     v.Sent,
					Ipaddr:   v.Ipaddr,
					Lastseen: v.Lastseen,
				})
		}
	} else {
		nodeList := hostlist.Expand(nodeNames)
		for _, v := range wwNodeStatus.Nodes {
			for j := 0; j < len(nodeList); j++ {
				if v.NodeName == nodeList[j] {
					nodeStatusResponse.NodeStatus = append(nodeStatusResponse.NodeStatus,
						&wwapiv1.NodeStatus{
							NodeName: v.NodeName,
							Stage:    v.Stage,
							Sent:     v.Sent,
							Ipaddr:   v.Ipaddr,
							Lastseen: v.Lastseen,
						})
					break
				}
			}
		}
	}
	return
}

// checkNetNameRequired is a helper for determining if netname is set.
// Certain settings require it.
func checkNetNameRequired(netname string) (err error) {
	if netname == "" {
		err = fmt.Errorf("You must include the '--netname' option")
		wwlog.Printf(wwlog.ERROR, fmt.Sprintf("%v\n", err.Error()))
	}
	return
}

// nodeDbSave persists the nodeDB to disk and restarts warewulfd.
// TODO: We will likely need locking around anything changing nodeDB
// or restarting warewulfd. Determine if the reason for restart is
// just to reinitialize warewulfd with the new nodeDB or if there is
// something more to it.
func nodeDbSave(nodeDB *node.NodeYaml) (err error) {
	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist nodedb")
	}

	err = warewulfd.DaemonReload()
	if err != nil {
		return errors.Wrap(err, "failed to reload warewulf daemon")
	}
	return
}
