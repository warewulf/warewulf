package apinode

import (
	"encoding/json"
	"fmt"
	"net/http"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

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

	controller := warewulfconf.Get()

	if controller.Ipaddr == "" {
		err = fmt.Errorf("the Warewulf Server IP Address is not properly configured")
		wwlog.Error(fmt.Sprintf("%v", err.Error()))
		return
	}

	statusURL := fmt.Sprintf("http://%s:%d/status", controller.Ipaddr, controller.Warewulf.Port)
	wwlog.Verbose("Connecting to: %s", statusURL)

	resp, err := http.Get(statusURL)
	if err != nil {
		wwlog.Error("Could not connect to Warewulf server: %s", err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var wwNodeStatus allStatus

	err = decoder.Decode(&wwNodeStatus)
	if err != nil {
		wwlog.Error("Could not decode JSON: %s", err)
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
