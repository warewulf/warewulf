package warewulfd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type nodeDB struct {
	lock     sync.RWMutex
	NodeInfo map[string]string
	yml      node.NodesYaml
}

var (
	db nodeDB
)

func LoadNodeDB() error {

	db.lock.Lock()
	defer db.lock.Unlock()
	return loadNodeDB()
}

func loadNodeDB() (err error) {
	TmpMap := make(map[string]string)

	db.yml, err = node.New()
	if err != nil {
		return
	}

	nodes, err := db.yml.FindAllNodes()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if n.Discoverable.Bool() {
			continue
		}
		for _, netdev := range n.NetDevs {
			hwaddr := strings.ToLower(netdev.Hwaddr)
			TmpMap[hwaddr] = n.Id()
		}
	}

	db.NodeInfo = TmpMap
	return nil
}

func GetNodeOrSetDiscoverable(hwaddr string) (node.Node, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	// NOTE: since discoverable nodes will write an updated DB to file and then
	// reload, it is not enough to lock individual reads from the DB
	// to ensure the condition on which the node is updated is still satisfied
	// after the DB is read back in.

	nId, ok := db.NodeInfo[hwaddr]
	if ok {
		return db.yml.GetNode(nId)
	}

	// If we failed to find a node, let's see if we can add one...
	wwlog.Warn("node not configured: %s", hwaddr)

	nodeFound, netdev, err := db.yml.FindDiscoverableNode()
	if err != nil {
		// NOTE: this is taken as there is no discoverable node, so return the
		// empty one
		return nodeFound, err
	}
	// update node
	wwlog.Debug("discovered node: %s netdev: %s", nodeFound.Id(), netdev)
	nodeChanges, _ := db.yml.GetNodeOnly(nodeFound.Id()) // ignore error as nodeId is in db
	if _, ok := nodeChanges.NetDevs[netdev]; !ok {
		nodeChanges.NetDevs = make(map[string]*node.NetDev)
		nodeChanges.NetDevs[netdev] = new(node.NetDev)
	}
	wwlog.Debug("node: %v", nodeChanges)
	nodeChanges.NetDevs[netdev].Hwaddr = hwaddr
	nodeChanges.Discoverable = "UNDEF"
	err = db.yml.SetNode(nodeFound.Id(), nodeChanges)
	if err != nil {
		return nodeFound, err
	}
	err = db.yml.Persist()
	if err != nil {
		return nodeFound, fmt.Errorf("%s (failed to persist node configuration) %w", hwaddr, err)
	}
	err = loadNodeDB()
	if err != nil {
		return nodeFound, fmt.Errorf("%s (failed to reload configuration) %w", hwaddr, err)
	}
	// NOTE: previously all overlays were built here, but that will also
	// be done automatically when attempting to serve an overlay that
	// hasn't been built (without blocking the database).

	wwlog.Serv("%s (node %s automatically configured)", hwaddr, nodeFound.Id())

	// return the discovered node
	return db.yml.GetNode(nodeFound.Id())
}

func Reload() {
	if err := LoadNodeDB(); err != nil {
		wwlog.Error("Could not load node DB: %s", err)
	}

	if err := LoadNodeStatus(); err != nil {
		wwlog.Error("Could not prepopulate node status DB: %s", err)
	}
}
