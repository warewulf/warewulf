package node

import (
	"errors"
	"net"
	"strings"
)

/*
get node by its hardware/MAC address, return error otherwise
*/
func (config *NodeYaml) FindByHwaddr(hwa string) (NodeInfo, error) {
	if _, err := net.ParseMAC(hwa); err != nil {
		return NodeInfo{}, errors.New("invalid hardware address: " + hwa)
	}

	var ret NodeInfo

	n, _ := config.FindAllNodes()

	for _, node := range n {
		for _, dev := range node.NetDevs {
			if strings.EqualFold(dev.Hwaddr.Get(), hwa) {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with HW Addr: " + hwa)
}

/*
get node by its ip address, return error otherwise
*/
func (config *NodeYaml) FindByIpaddr(ipaddr string) (NodeInfo, error) {
	if addr := net.ParseIP(ipaddr); addr == nil {
		return NodeInfo{}, errors.New("invalid IP:" + ipaddr)
	} else {
		ipaddr = addr.String()
	}
	var ret NodeInfo

	n, _ := config.FindAllNodes()

	for _, node := range n {
		for _, dev := range node.NetDevs {
			if dev.Ipaddr.Get() == ipaddr {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with IP Addr: " + ipaddr)
}

// return a single nodd identified by given id, returns error otherwise
func (config *NodeYaml) FindById(id string) (NodeInfo, error) {
	nodes, _ := config.FindAllNodes()
	for _, node := range nodes {
		if strings.EqualFold(node.Id.Get(), id) {
			return node, nil
		}
	}
	return NodeInfo{}, errors.New("no nodes found with id: " + id)
}

// Return just the node list as string slice
func (config *NodeYaml) NodeList() []string {
	ret := make([]string, len(config.Nodes))
	for key := range config.Nodes {
		ret = append(ret, key)
	}
	return ret
}
