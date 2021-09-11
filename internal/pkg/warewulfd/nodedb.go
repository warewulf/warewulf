package warewulfd

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/node"
)

type nodeDB struct {
	lock     sync.RWMutex
	NodeInfo map[string]node.NodeInfo
}

var (
	db nodeDB
)

func LoadNodeDB() error {
	TmpMap := make(map[string]node.NodeInfo)

	daemonLogf("(re)Loading the node Database\n")

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

		return db.NodeInfo[val], nil
	}

	var empty node.NodeInfo
	return empty, errors.New("No node found")
}
