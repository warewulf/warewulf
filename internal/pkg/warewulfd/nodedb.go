package warewulfd

import (
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

type nodeDB struct {
	lock     sync.RWMutex
	NodeInfo map[string]node.NodeInfo
}

var (
	db nodeDB
)

func LoadNodeDB() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	return loadNodeDB()
}

func loadNodeDB() error {
	TmpMap := make(map[string]node.NodeInfo)

	DB, err := node.New()
	if err != nil {
		return err
	}

	nodes, err := DB.FindAllNodes()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		for _, netdev := range n.NetDevs {
			hwaddr := strings.ToLower(netdev.Hwaddr.Get())
			TmpMap[hwaddr] = n
		}
	}

	db.NodeInfo = TmpMap

	return nil
}

func GetNode(val string) (node.NodeInfo, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return getNode(val)
}

func getNode(val string) (node.NodeInfo, error) {

	if _, ok := db.NodeInfo[val]; ok {

		return db.NodeInfo[val], nil
	}

	var empty node.NodeInfo
	return empty, errors.New("No node found")
}

func GetNodeOrSetDiscoverable(hwaddr string) (node.NodeInfo, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	return getNodeOrSetDiscoverable(hwaddr)
}

func getNodeOrSetDiscoverable(hwaddr string) (node.NodeInfo, error) {
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

	config, err := node.New()
	if err != nil {
		return n, errors.Wrapf(err, "%s (failed to read node configuration file)", hwaddr)
	}

	_n, netdev, err := config.FindDiscoverableNode()
	if err != nil {
		// NOTE: this is taken as there is no discoverable node, so return the
		// empty one
		return n, nil
	}

	_n.NetDevs[netdev].Hwaddr.Set(hwaddr)
	_n.Discoverable.SetB(false)

	// NOTE: errors here should return the empty node if the state cannot
	// be saved and re-loaded, since subsequent requests will be made on invalid
	// assumption that the database is up to date.
	err = config.NodeUpdate(_n)
	if err != nil {
		return n, errors.Wrapf(err, "%s (failed to set node configuration)", hwaddr)
	}

	err = config.Persist()
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
	return _n, nil
}
