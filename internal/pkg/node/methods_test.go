package node

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IpCIDR(t *testing.T) {
	tests := map[string]struct {
		ipaddr  net.IP
		netmask net.IP
		cidr    string
	}{
		"nil": {
			ipaddr:  nil,
			netmask: nil,
			cidr:    "",
		},
		"ip only": {
			ipaddr:  net.ParseIP("192.168.1.1"),
			netmask: nil,
			cidr:    "",
		},
		"netmask only": {
			ipaddr:  nil,
			netmask: net.ParseIP("255.255.255.0"),
			cidr:    "",
		},
		"working": {
			ipaddr:  net.ParseIP("192.168.1.1"),
			netmask: net.ParseIP("255.255.255.0"),
			cidr:    "192.168.1.1/24",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			n := new(NetDev)
			n.Ipaddr = tt.ipaddr
			n.Netmask = tt.netmask
			assert.Equal(t, tt.cidr, n.IpCIDR())
		})
	}
}

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

func Test_Node_Expand_Flatten(t *testing.T) {
	node := new(Node)

	assert.Nil(t, node.Ipmi)
	assert.Nil(t, node.Kernel)
	assert.Nil(t, node.NetDevs)
	assert.Nil(t, node.Tags)

	t.Run("test expand", func(t *testing.T) {
		node.Expand()
		assert.Equal(t, map[string]string{}, node.Tags)
		assert.Equal(t, map[string]string{}, node.Ipmi.Tags)
		assert.Equal(t, "", node.Kernel.Version)
		assert.Len(t, node.Kernel.Args, 0)
		assert.Equal(t, map[string]*NetDev{}, node.NetDevs)
	})

	t.Run("test flatten", func(t *testing.T) {
		node.Flatten()
		assert.Nil(t, node.Ipmi)
		assert.Nil(t, node.Kernel)
		assert.Nil(t, node.NetDevs)
		assert.Nil(t, node.Tags)
	})
}

func Test_Profile_Expand_Flatten(t *testing.T) {
	profile := new(Profile)

	assert.Nil(t, profile.Ipmi)
	assert.Nil(t, profile.Kernel)
	assert.Nil(t, profile.NetDevs)
	assert.Nil(t, profile.Tags)

	t.Run("test expand", func(t *testing.T) {
		profile.Expand()
		assert.Equal(t, map[string]string{}, profile.Tags)
		assert.Equal(t, map[string]string{}, profile.Ipmi.Tags)
		assert.Equal(t, "", profile.Kernel.Version)
		assert.Len(t, profile.Kernel.Args, 0)
		assert.Equal(t, map[string]*NetDev{}, profile.NetDevs)
	})

	t.Run("test flatten", func(t *testing.T) {
		profile.Flatten()
		assert.Nil(t, profile.Ipmi)
		assert.Nil(t, profile.Kernel)
		assert.Nil(t, profile.NetDevs)
		assert.Nil(t, profile.Tags)
	})
}
