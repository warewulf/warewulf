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
	NodeInfo map[string]node.NodeConf
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
	TmpMap := make(map[string]node.NodeConf)

	db.yml, err = node.New()
	if err != nil {
		return
	}

	nodes, err := db.yml.FindAllNodes()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		if n.Discoverable {
			continue
		}
		for _, netdev := range n.NetDevs {
			hwaddr := strings.ToLower(netdev.Hwaddr)
			TmpMap[hwaddr] = n
		}
	}

	db.NodeInfo = TmpMap
	return nil
}

func GetNode(val string) (node node.NodeConf, err error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return getNode(val)
}

func getNode(val string) (node node.NodeConf, err error) {

	if _, ok := db.NodeInfo[val]; ok {

		return db.NodeInfo[val], nil
	}

	return node, errors.New("No node found")
}

func GetNodeOrSetDiscoverable(hwaddr string) (node.NodeConf, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	return getNodeOrSetDiscoverable(hwaddr)
}

func PersistDb() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	return db.yml.Persist()
}

func getNodeOrSetDiscoverable(hwaddr string) (node.NodeConf, error) {
	// NOTE: since discoverable nodes will write an updated DB to file and then
	// reload, it is not enough to lock individual reads from the DB
	// to ensure the condition on which the node is updated is still satisfied
	// after the DB is read back in.

	n, err := getNode(hwaddr)
	if err == nil {
		return n, nil
	}

	// If we failed to find a node, let's see if we can add one...
	var netdev string

	wwlog.WarnExc(err, "%s (node not configured)", hwaddr)

	nodeId, netdev, err := db.yml.FindDiscoverableNode()
	if err != nil {
		// NOTE: this is taken as there is no discoverable node, so return the
		// empty one
		return n, nil
	}
	discoverdNode, err := getNode(nodeId)
	if err != nil {
		return n, err
	}
	// update data on disk and in memory
	discoverdNode.NetDevs[netdev].Hwaddr = hwaddr
	db.yml.Nodes[discoverdNode.Id()].NetDevs[netdev].Hwaddr = hwaddr
	discoverdNode.Discoverable = false
	db.yml.Nodes[discoverdNode.Id()].Discoverable = false

	err = PersistDb()
	if err != nil {
		return n, errors.Wrapf(err, "%s (failed to persist node configuration)", hwaddr)
	}

	err = loadNodeDB()
	if err != nil {
		return n, errors.Wrapf(err, "%s (failed to reload configuration)", hwaddr)
	}

	// NOTE: previously all overlays were built here, but that will also
	// be done automatically when attempting to serve an overlay that
	// hasn't been built (without blocking the database).

	wwlog.Serv("%s (node automatically configured)", hwaddr)

	// return the discovered node
	return discoverdNode, nil
}
