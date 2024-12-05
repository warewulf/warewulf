package warewulfd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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

var (
	statusDB allStatus
	dbLock   = sync.RWMutex{}
)

func init() {
	statusDB.Nodes = make(map[string]*NodeStatus)
}

func LoadNodeStatus() error {
	dbLock.Lock()
	defer dbLock.Unlock()
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
		if _, ok := statusDB.Nodes[n.Id()]; !ok {
			newDB.Nodes[n.Id()] = &NodeStatus{}
			newDB.Nodes[n.Id()].NodeName = n.Id()
		} else {
			newDB.Nodes[n.Id()] = statusDB.Nodes[n.Id()]
		}
	}

	statusDB = newDB
	return nil
}

func updateStatus(nodeID, stage, sent, ipaddr string) {
	dbLock.Lock()
	defer dbLock.Unlock()
	rightnow := time.Now().Unix()

	wwlog.Debug("Updating node status data: %s", nodeID)

	n := NodeStatus{
		NodeName: nodeID,
		Stage:    stage,
		Lastseen: rightnow,
		Sent:     sent,
		Ipaddr:   ipaddr,
	}
	statusDB.Nodes[nodeID] = &n
}

func statusJSON() ([]byte, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	wwlog.Debug("Request for node status data...")

	ret, err := json.MarshalIndent(statusDB, "", "  ")
	if err != nil {
		return ret, fmt.Errorf("could not marshal JSON data from status structure: %w", err)
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
