package node

import (
	"errors"
	"net"
	"strings"
)

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
