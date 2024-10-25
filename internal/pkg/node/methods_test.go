package node

import (
	"net"
	"testing"
)

func Test_Empty(t *testing.T) {
	var netdev NetDev
	var netdevPtr *NetDev

	t.Run("test for empty", func(t *testing.T) {
		if ObjectIsEmpty(netdev) != true {
			t.Errorf("netdev must be empty")
		}
	})
	t.Run("test for non empty", func(t *testing.T) {
		netdev.Device = "foo"
		if ObjectIsEmpty(netdev) == true {
			t.Errorf("netdev must be non empty")
		}
	})
	t.Run("test for nil pointer", func(t *testing.T) {
		if ObjectIsEmpty(netdevPtr) != true {
			t.Errorf("netdev must be empty")
		}
	})
	t.Run("test for pointer assigned", func(t *testing.T) {
		netdev.Ipaddr = net.ParseIP("10.10.10.1")
		netdevPtr = &netdev
		if ObjectIsEmpty(netdevPtr) == true {
			t.Errorf("netdev must be empty")
		}
	})
}
