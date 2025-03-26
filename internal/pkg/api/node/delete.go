package apinode

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// NodeDelete adds nodes for management by Warewulf.
func NodeDelete(ndp *wwapiv1.NodeDeleteParameter) (err error) {

	var nodeList []node.Node
	nodeList, err = NodeDeleteParameterCheck(ndp, false)
	if err != nil {
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Failed to open node database: %s", err)
		return
	}
	dbHash := nodeDB.Hash()
	if hex.EncodeToString(dbHash[:]) != ndp.Hash && !ndp.Force {
		return fmt.Errorf("got wrong hash, not modifying node database")
	}

	for _, n := range nodeList {
		err := nodeDB.DelNode(n.Id())
		if err != nil {
			wwlog.Error("%s", err)
		} else {
			wwlog.Verbose("Deleting node: %s\n", n.Id())
		}
	}

	err = nodeDB.Persist()
	if err != nil {
		return fmt.Errorf("failed to persist nodedb: %w", err)
	}

	err = daemon.DaemonReload()
	if err != nil {
		return fmt.Errorf("failed to reload warewulf daemon: %w", err)
	}
	return
}

// NodeDeleteParameterCheck does error checking on NodeDeleteParameter.
// Output to the console if console is true.
// Returns the nodes to delete.
func NodeDeleteParameterCheck(ndp *wwapiv1.NodeDeleteParameter, console bool) (nodeList []node.Node, err error) {

	if ndp == nil {
		err = fmt.Errorf("NodeDeleteParameter is nil")
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Failed to open node database: %s", err)
		return
	}
	dbHash := nodeDB.Hash()
	if hex.EncodeToString(dbHash[:]) != ndp.Hash && !ndp.Force {
		wwlog.Debug("got hash: %s", ndp.Hash)
		wwlog.Debug("actual hash: %s", hex.EncodeToString(dbHash[:]))
		err = fmt.Errorf("got wrong hash, not modifying node database")
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("Could not get node list: %s", err)
		return
	}

	node_args := hostlist.Expand(ndp.NodeNames)

	for _, r := range node_args {
		var match bool
		for _, n := range nodes {
			if n.Id() == r {
				nodeList = append(nodeList, n)
				match = true
			}
		}

		if !match {
			fmt.Fprintf(os.Stderr, "ERROR: No match for node: %s\n", r)
		}
	}

	if len(nodeList) == 0 {
		fmt.Printf("No nodes found\n")
	}
	return
}
