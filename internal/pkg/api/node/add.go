package apinode

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// NodeAdd adds nodes for management by Warewulf.
func NodeAdd(nap *wwapiv1.NodeAddParameter) (err error) {

	if nap == nil {
		return fmt.Errorf("NodeAddParameter is nil")
	}

	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "failed to open node database")
	}
	dbHash := nodeDB.Hash()
	if hex.EncodeToString(dbHash[:]) != nap.Hash && !nap.Force {
		return fmt.Errorf("got wrong hash, not modifying node database")
	}
	node_args := hostlist.Expand(nap.NodeNames)
	var ipv4, ipmiaddr net.IP
	for _, a := range node_args {
		n, err := nodeDB.AddNode(a)
		if err != nil {
			return errors.Wrap(err, "failed to add node")
		}
		err = yaml.Unmarshal([]byte(nap.NodeConfYaml), &n)
		if err != nil {
			return errors.Wrap(err, "Failed to decode nodeConf")
		}
		wwlog.Info("Added node: %s:", a)
		for _, dev := range n.NetDevs {
			if !ipv4.IsUnspecified() && ipv4 != nil {
				// if more nodes are added increment IPv4 address
				ipv4 = util.IncrementIPv4(ipv4, 1)
				wwlog.Verbose("Incremented IP addr to %s", ipv4)
				dev.Ipaddr = ipv4

			} else if !dev.Ipaddr.IsUnspecified() {
				ipv4 = dev.Ipaddr
			}
		}
		if n.Ipmi != nil {
			if !ipmiaddr.IsUnspecified() && ipmiaddr != nil {
				ipmiaddr = util.IncrementIPv4(ipmiaddr, 1)
				wwlog.Verbose("Incremented ipmi IP addr to %s", ipmiaddr)
				n.Ipmi.Ipaddr = ipmiaddr
			} else if !n.Ipmi.Ipaddr.IsUnspecified() {
				ipmiaddr = n.Ipmi.Ipaddr
			}
		}
	}

	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist new node")
	}

	err = warewulfd.DaemonReload()
	if err != nil {
		return errors.Wrap(err, "failed to reload warewulf daemon")
	}
	return
}
