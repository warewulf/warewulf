package warewulfd

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "warewulf_status"
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

var prthLabels = []string{"nodeID", "stage", "sent", "ipaddr"}

type containerSize struct {
	size int
	name string
}
type containerSizes struct {
	sync.Mutex
	containers map[string]*containerSize
}
type Collector struct {
	sync.Mutex
	numNodes      prometheus.Gauge
	lastseen      *prometheus.GaugeVec
	containerSize *prometheus.GaugeVec
}

var statusDB allStatus

var containerDB containerSizes

func init() {
	statusDB.Nodes = make(map[string]*NodeStatus)
	containerDB.containers = make(map[string]*containerSize)
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

/*
Update the container size for the given size or the given
containers.
*/
func UpdateContainerSize(cntName ...string) {
	containerDB.Lock()
	defer containerDB.Unlock()
	var sources []string
	if len(cntName) != 0 {
		sources = append(sources, cntName...)
	} else {
		sources, _ = container.ListSources()
	}

	for _, source := range sources {
		srcSize, _ := util.DirSize(container.SourceDir(source))
		imgSize, _ := os.Stat(container.ImageFile(source))
		containerDB.containers[source] = &containerSize{
			size: int(srcSize) + int(imgSize.Size()),
			name: source,
		}
	}
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

func NewCollector() *Collector {
	return &Collector{
		numNodes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "num_nodes",
				Help:      "Number of nodes",
			},
		),
		lastseen: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "node_lastseen",
				Help:      "Last time in seconds when the node was last seen in stage",
			},
			prthLabels,
		),
		containerSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "container_size",
				Help:      "Size of the container including the compressed image",
			},
			[]string{"container"},
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.numNodes.Desc()
	c.lastseen.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Lock()
	defer c.Unlock()
	c.lastseen.Reset()
	c.containerSize.Reset()
	c.numNodes.Set(float64(len(statusDB.Nodes)))
	ch <- c.numNodes
	for _, n := range statusDB.Nodes {
		c.lastseen.WithLabelValues(n.NodeName, n.Stage, n.Sent, n.Ipaddr).Set(float64(n.Lastseen))
	}
	c.lastseen.Collect(ch)
	for _, cnt := range containerDB.containers {
		c.containerSize.WithLabelValues(cnt.name).Set(float64(cnt.size))
	}
	c.containerSize.Collect(ch)
}
