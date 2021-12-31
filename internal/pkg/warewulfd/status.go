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
	Stage    string `json:"stage"`
	Sent     string `json:"sent"`
	Ipaddr   string `json:"ipaddr"`
	Lastseen int64  `json:"last seen"`
}

var statusDB allStatus

func init() {
	statusDB.Nodes = make(map[string]*NodeStatus)

	err := LoadNodeStatus()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not prepopulate status DB with nodes: %s\n", err)
	}
}

func LoadNodeStatus() error {
	var newDB allStatus
	newDB.Nodes = make(map[string]*NodeStatus)

	DB, err := node.New()
	if err != nil {
		return err
	}

	nodes, err := DB.FindAllNodes()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if _, ok := statusDB.Nodes[n.Id.Get()]; !ok {
			newDB.Nodes[n.Id.Get()] = &NodeStatus{}
		} else {
			newDB.Nodes[n.Id.Get()] = statusDB.Nodes[n.Id.Get()]
		}
	}

	statusDB = newDB
	return nil
}

func updateStatus(nodeID, stage, sent, ipaddr string) {
	rightnow := time.Now().Unix()

	wwlog.Printf(wwlog.DEBUG, "Updating node status data: %s\n", nodeID)

	var n NodeStatus
	n.Stage = stage
	n.Lastseen = rightnow
	n.Sent = sent
	n.Ipaddr = ipaddr
	statusDB.Nodes[nodeID] = &n
}

func statusJSON() ([]byte, error) {

	wwlog.Printf(wwlog.DEBUG, "Request for node status data...\n")

	ret, err := json.MarshalIndent(statusDB, "", "  ")
	if err != nil {
		return ret, errors.Wrap(err, "could not marshal JSON data from sstatus structure")
	}

	return ret, nil
}

func StatusSend(w http.ResponseWriter, req *http.Request) {

	status, err := statusJSON()
	if err != nil {
		return
	}

	_, err = w.Write(status)
	if err != nil {
		wwlog.Printf(wwlog.WARN, "Could not send status JSON: %s\n", err)
	}
}
