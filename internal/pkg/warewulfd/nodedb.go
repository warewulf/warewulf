package warewulfd

import (
	"sync"

	"github.com/hpcng/warewulf/internal/pkg/errors"
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

	var TmpMap map[string]node.NodeInfo
	TmpMap = make(map[string]node.NodeInfo)

	wwlog.Printf(wwlog.INFO, "Loading the node Database\n")

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
			wwlog.Printf(wwlog.DEBUG, "Caching node entry: '%s' -> %s\n", netdev.Hwaddr.Get(), n.Id.Get())

			TmpMap[netdev.Hwaddr.Get()] = n
		}
	}

	db.lock.Lock()
	defer db.lock.Unlock()
	db.NodeInfo = TmpMap

	return nil
}

func GetNode(val string) (node.NodeInfo, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if _, ok := db.NodeInfo[val]; ok {
		wwlog.Printf(wwlog.DEBUG, "Found node:\n%+v\n", db.NodeInfo[val])

		return db.NodeInfo[val], nil
	}

	wwlog.Printf(wwlog.VERBOSE, "Node not found in DB: %s\n", val)
	var empty node.NodeInfo
	return empty, errors.New("No node found")
}
