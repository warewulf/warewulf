package warewulfd

import (
	"encoding/json"
	"net/http"
	"time"

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
}

func updateStatus(nodeID, stage, sent, ipaddr string) error {
	rightnow := time.Now().Unix()

	wwlog.Printf(wwlog.DEBUG, "Updating node status data: %s\n", nodeID)

	var n NodeStatus
	n.Stage = stage
	n.Lastseen = rightnow
	n.Sent = sent
	n.Ipaddr = ipaddr
	statusDB.Nodes[nodeID] = &n

	return nil
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

	w.Write(status)
}
