package warewulfd

import (
	"fmt"
	"strings"
	"sync"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
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

// GetNode looks up a configured node by hardware address. It returns an error
// if no node is configured for the given hwaddr.
func GetNode(hwaddr string) (node.Node, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	nId, ok := db.NodeInfo[hwaddr]
	if !ok {
		return node.Node{}, fmt.Errorf("no node configured for hwaddr %s", hwaddr)
	}
	return db.yml.GetNode(nId)
}

func GetOrDiscoverNode(hwaddr string, autobuildOverlays bool) (node.Node, error) {
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
	if autobuildOverlays {
		if err := overlay.RemoveImage(nodeFound.Id(), "system", []string{}); err != nil {
			wwlog.Warn("Failed to clear system overlay image: %s: %s", nodeFound.Id(), err)
		}
		if err := overlay.RemoveImage(nodeFound.Id(), "runtime", []string{}); err != nil {
			wwlog.Warn("Failed to clear runtime overlay image: %s: %s", nodeFound.Id(), err)
		}
	}

	wwlog.Serv("%s (node %s automatically configured)", hwaddr, nodeFound.Id())

	// return the discovered node
	return db.yml.GetNode(nodeFound.Id())
}

// usesTwoStageBoot reports whether n boots two-stage with dracut, which
// can fetch the system overlay from a privileged port. ok is false when
// n uses a custom iPXE template, whose boot method is unknown.
func usesTwoStageBoot(n node.Node, grubBoot bool) (twoStage bool, ok bool) {
	if grubBoot {
		return n.Tags["GrubMenuEntry"] == "dracut", true
	}
	if n.Ipxe != "" && n.Ipxe != "default" {
		return false, false
	}
	return strings.HasPrefix(n.Tags["IPXEMenuEntry"], "dracut"), true
}

// warnSingleStageSecureSystem warns when warewulf:secure includes
// "system" and some nodes do not boot two-stage, so they cannot fetch
// the system overlay.
func warnSingleStageSecureSystem() {
	conf := warewulfconf.Get()
	if !conf.Warewulf.SecureSystemOverlay() {
		return
	}
	db.lock.RLock()
	nodes, err := db.yml.FindAllNodes()
	db.lock.RUnlock()
	if err != nil {
		return
	}
	var ids []string
	for _, n := range nodes {
		if twoStage, ok := usesTwoStageBoot(n, conf.Warewulf.GrubBoot()); ok && !twoStage {
			ids = append(ids, n.Id())
		}
	}
	if len(ids) == 0 {
		return
	}
	tag := "IPXEMenuEntry"
	if conf.Warewulf.GrubBoot() {
		tag = "GrubMenuEntry"
	}
	wwlog.Warn("warewulf:secure includes \"system\", but these nodes do not boot two-stage with dracut and cannot fetch the system overlay: %s (set tag %s=dracut)",
		strings.Join(ids, ", "), tag)
}

func Reload() {
	if err := LoadNodeDB(); err != nil {
		wwlog.Error("Could not load node DB: %s", err)
	} else {
		warnSingleStageSecureSystem()
	}

	if err := LoadNodeStatus(); err != nil {
		wwlog.Error("Could not prepopulate node status DB: %s", err)
	}
}
