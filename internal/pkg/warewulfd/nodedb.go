package warewulfd

import (
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type nodeDB struct {
	lock     sync.RWMutex
	NodeInfo map[string]string
	yml      node.NodeYaml
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

func GetNodeOrSetDiscoverable(hwaddr string) (node.NodeConf, error) {
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

	node, netdev, err := db.yml.FindDiscoverableNode()
	if err != nil {
		// NOTE: this is taken as there is no discoverable node, so return the
		// empty one
		return node, err
	}
	// update node
	wwlog.Debug("discoverd node: %s netdev: %s", node.Id(), netdev)
	nodeChanges, _ := db.yml.GetNodeOnly(node.Id()) // ignore error as nodeId is in db
	wwlog.Debug("node: %v", nodeChanges)
	nodeChanges.NetDevs[netdev].Hwaddr = hwaddr
	nodeChanges.Discoverable = "UNDEF"
	err = db.yml.SetNode(node.Id(), nodeChanges)
	if err != nil {
		return node, err
	}
	err = db.yml.Persist()
	if err != nil {
		return node, errors.Wrapf(err, "%s (failed to persist node configuration)", hwaddr)
	}
	err = loadNodeDB()
	if err != nil {
		return node, errors.Wrapf(err, "%s (failed to reload configuration)", hwaddr)
	}
	// NOTE: previously all overlays were built here, but that will also
	// be done automatically when attempting to serve an overlay that
	// hasn't been built (without blocking the database).

	wwlog.Serv("%s (node %s automatically configured)", hwaddr, node.Id())

	// return the discovered node
	return db.yml.GetNode(node.Id())
}
