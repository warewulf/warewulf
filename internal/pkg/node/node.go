// Package node implements data structures and functions for nodes
// (and, by extension, profiles).
//
// The *Conf types (NodeConf, IpmiConf, KernelConf, NetDevs) represent
// the literal, un-implemented configuration as it would appear in
// nodes.conf.
//
// The *Info and *Entry types (NodeInfo, IpmiEntry, KernelEntry,
// NetDevEntry) represent the logical, effective configuration as
// interpreted from the configuration. For example, values inherited
// from a profile are visible on a node's NodeInfo.
package node

import (
	"path"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
)


var ConfigFile string
var DefaultConfig string

// used as fallback if DefaultConfig can't be read
var FallBackConf = `---
defaultnode:
  runtime overlay:
  - generic
  system overlay:
  - wwinit
  kernel:
    args: quiet crashkernel=no vga=791 net.naming-scheme=v238
  init: /sbin/init
  root: initramfs
  ipxe template: default
  profiles:
  - default
  network devices:
    dummy:
      device: eth0
      type: ethernet
      netmask: 255.255.255.0`


func init() {
	if ConfigFile == "" {
		ConfigFile = path.Join(buildconfig.SYSCONFDIR(), "warewulf/nodes.conf")
	}
	if DefaultConfig == "" {
		DefaultConfig = path.Join(buildconfig.SYSCONFDIR(), "warewulf/defaults.conf")
	}
}
