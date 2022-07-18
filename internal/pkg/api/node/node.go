package apinode

import (
	"encoding/json"
	"fmt"
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
		if nap.OptionsStrMap["ipaddr"] != "" {
			// if more nodes are added increment IPv4 address
			nap.OptionsStrMap["ipaddr"] = util.IncrementIPv4(nap.OptionsStrMap["ipaddr"], count)

			wwlog.Verbose("Node: %s:%s, Setting Ipaddr to: %s\n",
				n.Id.Get(), nap.OptionsStrMap["netname"], nap.OptionsStrMap["ipaddr"])
		}
		if nap.OptionsStrMap["ipmiaddr"] != "" {
			// if more nodes are added increment IPv4 address
			nap.OptionsStrMap["ipmiaddr"] = util.IncrementIPv4(nap.OptionsStrMap["ipmiaddr"], count)

			wwlog.Verbose("Node: %s:, Setting IPMIIpaddr to: %s\n",
				n.Id.Get(), nap.OptionsStrMap["ipmiaddr"])
		}
		// Now set all the rest
		for key, val := range nap.OptionsStrMap {
			if val != "" {
				wwlog.Verbose("node:%s setting %s to %s\n", n.Id.Get(), key, val)
				n.SetField(key, val)
			}
		}

		err = nodeDB.NodeUpdate(n)
		if err != nil {
			return errors.Wrap(err, "failed to update nodedb")
		}

		count++
	}

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
		ni.Tags = map[string]*wwapiv1.NodeField{}
		for keyname, keyvalue := range node.Tags {
			ni.Tags[keyname] = new(wwapiv1.NodeField)
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
	return DbSave(&nodeDB)
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
		for key, val := range set.OptionsStrMap {
			if val != "" {
				wwlog.Verbose("node:%s setting %s to %s\n", n.Id.Get(), key, val)
				n.SetField(key, val)
			}
		}

		if set.NetdevDelete != "" {

			if _, ok := n.NetDevs[set.NetdevDelete]; !ok {
				err = fmt.Errorf("Network device name doesn't exist: %s", set.NetdevDelete)
				wwlog.Printf(wwlog.ERROR, fmt.Sprintf("%v\n", err.Error()))
				return
			}

			wwlog.Printf(wwlog.VERBOSE, "Node: %s, Deleting network device: %s\n", n.Id.Get(), set.NetdevDelete)
			delete(n.NetDevs, set.NetdevDelete)
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

/*
Add the netname to the options map, as its only known after the map
command line options have been read out.
*/
func AddNetname(theMap map[string]*string) (map[string]*string, bool) {
	foundNetname := false
	netname := ""
	retMap := make(map[string]*string)
	for key, val := range theMap {
		if key == "NetDevs" {
			foundNetname = true
			netname = *val
		}
	}
	if foundNetname {
		for key, val := range theMap {
			keys := strings.Split(key, ".")
			myVal := *val
			if len(keys) >= 2 && keys[0] == "NetDevs" {
				retMap[keys[0]+"."+netname+"."+strings.Join(keys[1:], ".")] = &myVal
			} else {
				retMap[key] = &myVal
			}
		}
	}
	return retMap, foundNetname
}
