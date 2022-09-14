package apinode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
)

// NodeAdd adds nodes for management by Warewulf.
func NodeAdd(nap *wwapiv1.NodeAddParameter) (err error) {

	if nap == nil {
		return fmt.Errorf("NodeAddParameter is nil")
	}

	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "failed to open node database")
	}

	node_args := hostlist.Expand(nap.NodeNames)
	var nodeConf node.NodeConf
	err = yaml.Unmarshal([]byte(nap.NodeConfYaml), &nodeConf)
	if err != nil {
		return errors.Wrap(err, "Failed to decode nodeConf")
	}

	for _, a := range node_args {
		var n node.NodeInfo
		n, err = nodeDB.AddNode(a)
		if err != nil {
			return errors.Wrap(err, "failed to add node")
		}
		wwlog.Info("Added node: %s\n", a)
		var netName string
		for netName = range nodeConf.NetDevs {
			// as map should only have key this should give is the first and
			// only key
		}
		// setting node from the received yaml
		n.SetFrom(&nodeConf)
		if netName != "" && nodeConf.NetDevs[netName].Ipaddr != "" {
			// if more nodes are added increment IPv4 address
			nodeConf.NetDevs[netName].Ipaddr = util.IncrementIPv4(nodeConf.NetDevs[netName].Ipaddr, 1)

			wwlog.Verbose("Incremented IP addr to %s\n", nodeConf.NetDevs[netName].Ipaddr)
		}
		if nodeConf.Ipmi != nil && nodeConf.Ipmi.Ipaddr != "" {
			// if more nodes are added increment IPv4 address
			nodeConf.Ipmi.Ipaddr = util.IncrementIPv4(nodeConf.Ipmi.Ipaddr, 1)
			wwlog.Verbose("Incremented IP addr to %s\n", nodeConf.Ipmi.Ipaddr)
		}
		err = nodeDB.NodeUpdate(n)
		if err != nil {
			return errors.Wrap(err, "failed to update nodedb")
		}

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
		wwlog.Error("Failed to open node database: %s\n", err)
		return
	}

	for _, n := range nodeList {
		err := nodeDB.DelNode(n.Id.Get())
		if err != nil {
			wwlog.Error("%s\n", err)
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
		wwlog.Error("Failed to open node database: %s\n", err)
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("Could not get node list: %s\n", err)
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
		err = fmt.Errorf("node set parameter is nil")
		if console {
			fmt.Printf("%v\n", err)
			return
		}
	}

	if set.NodeNames == nil {
		err = fmt.Errorf("node set parameter: NodeNames is nil")
		if console {
			fmt.Printf("%v\n", err)
			return
		}
	}

	nodeDB, err = node.New()
	if err != nil {
		wwlog.Error("Could not open node configuration: %s\n", err)
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("Could not get node list: %s\n", err)
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
		wwlog.Verbose("Evaluating node: %s\n", n.Id.Get())
		var nodeConf node.NodeConf
		err = yaml.Unmarshal([]byte(set.NodeConfYaml), &nodeConf)
		if err != nil {
			wwlog.Error(fmt.Sprintf("%v\n", err.Error()))
			return
		}
		n.SetFrom(&nodeConf)
		if set.NetdevDelete != "" {
			if _, ok := n.NetDevs[set.NetdevDelete]; !ok {
				err = fmt.Errorf("network device name doesn't exist: %s", set.NetdevDelete)
				wwlog.Error(fmt.Sprintf("%v\n", err.Error()))
				return
			}

			wwlog.Verbose("Node: %s, Deleting network device: %s\n", n.Id.Get(), set.NetdevDelete)
			delete(n.NetDevs, set.NetdevDelete)
		}
		for _, key := range nodeConf.TagsDel {
			delete(n.Tags, key)
		}
		for _, key := range nodeConf.Ipmi.TagsDel {
			delete(n.Ipmi.Tags, key)
		}
		for net := range nodeConf.NetDevs {
			for _, key := range nodeConf.NetDevs[net].TagsDel {
				if _, ok := n.NetDevs[net]; ok {
					delete(n.NetDevs[net].Tags, key)
				}
			}
		}
		err := nodeDB.NodeUpdate(n)
		if err != nil {
			wwlog.Error("%s\n", err)
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
		wwlog.Error("%s\n", err)
		return
	}

	if controller.Ipaddr == "" {
		err = fmt.Errorf("the Warewulf Server IP Address is not properly configured")
		wwlog.Error(fmt.Sprintf("%v\n", err.Error()))
		return
	}

	statusURL := fmt.Sprintf("http://%s:%d/status", controller.Ipaddr, controller.Warewulf.Port)
	wwlog.Verbose("Connecting to: %s\n", statusURL)

	resp, err := http.Get(statusURL)
	if err != nil {
		wwlog.Error("Could not connect to Warewulf server: %s\n", err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var wwNodeStatus allStatus

	err = decoder.Decode(&wwNodeStatus)
	if err != nil {
		wwlog.Error("Could not decode JSON: %s\n", err)
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
