package warewulfd

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

type allStatus struct {
	Nodes map[string]*NodeStatus `json:"nodes"`
}

type NodeStatus struct {
	NodeName string `json:"node name"`
	Stage    string `json:"stage"`
	Sent     string `json:"sent"`
	Ipaddr   string `json:"ipaddr"`
	Lastseen int64  `json:"last seen"`
}

var statusDB allStatus

func init() {
	statusDB.Nodes = make(map[string]*NodeStatus)
}

func LoadNodeStatus() error {
	var newDB allStatus
	newDB.Nodes = make(map[string]*NodeStatus)

	DB, err := node.ReadNodeYaml()
	if err != nil {
		return err
	}

	nodes, err := DB.GetAllNodeInfo()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if _, ok := statusDB.Nodes[n.Id.Get()]; !ok {
			newDB.Nodes[n.Id.Get()] = &NodeStatus{}
			newDB.Nodes[n.Id.Get()].NodeName = n.Id.Get()
		} else {
			newDB.Nodes[n.Id.Get()] = statusDB.Nodes[n.Id.Get()]
		}
	}

	statusDB = newDB
	return nil
}

func updateStatus(nodeID, stage, sent, ipaddr string) {
	rightnow := time.Now().Unix()

	wwlog.Debug("Updating node status data: %s", nodeID)

	var n NodeStatus
	n.NodeName = nodeID
	n.Stage = stage
	n.Lastseen = rightnow
	n.Sent = sent
	n.Ipaddr = ipaddr
	statusDB.Nodes[nodeID] = &n
}

func statusJSON() ([]byte, error) {

	wwlog.Debug("Request for node status data...")

	ret, err := json.MarshalIndent(statusDB, "", "  ")
	if err != nil {
		return ret, errors.Wrap(err, "could not marshal JSON data from sstatus structure")
	}

	return ret, nil
}

func StatusSend(w http.ResponseWriter, req *http.Request) {

	status, err := statusJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(status)
	if err != nil {
		wwlog.Warn("Could not send status JSON: %s", err)
	}
}
