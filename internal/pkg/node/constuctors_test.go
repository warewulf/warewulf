package node

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func newConstructorPrimaryNetworkTest() NodeYaml {
	var data = `
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
nodes:
  test_node1:
    network devices:
      net0:
        device: eth0
  test_node2:
    primary network: net1
    network devices:
      net0:
        device: eth0
      net1:
        device: eth1
  test_node3:
    network devices:
      net0:
        device: eth0
      net1:
        device: eth1
  test_node4:
    primary network: net3
    network devices:
      net0:
        device: eth0
      net1:
        device: eth1
  `
	var ret NodeYaml
	_ = yaml.Unmarshal([]byte(data), &ret)
	return ret
}
func Test_Primary_Network(t *testing.T) {
	c := newConstructorPrimaryNetworkTest()
	nodes, _ := c.FindAllNodes()
	test_node1 := NewInfo()
	test_node2 := NewInfo()
	test_node3 := NewInfo()
	test_node4 := NewInfo()

	for _, n := range nodes {
		if n.Id.Get() == "test_node1" {
			test_node1 = n
		}
		if n.Id.Get() == "test_node2" {
			test_node2 = n
		}
		if n.Id.Get() == "test_node3" {
			test_node3 = n
		}
		if n.Id.Get() == "test_node4" {
			test_node4 = n
		}
	}
	t.Run("Primary network with one network, nothing set", func(t *testing.T) {
		if test_node1.PrimaryNetDev.Get() != "net0" {
			t.Errorf("primary network isn't net0 but: %s", test_node1.PrimaryNetDev.Get())
		}
		if !test_node1.NetDevs["net0"].Primary.GetB() {
			t.Errorf("primary flag isn't set for net0")
		}
	})
	t.Run("Primary network with two networks, primary is net1", func(t *testing.T) {
		if test_node2.PrimaryNetDev.Get() != "net1" {
			t.Errorf("primary network isn't net1 but: %s", test_node2.PrimaryNetDev.Get())
		}
		if test_node2.NetDevs["net0"].Primary.GetB() {
			t.Errorf("primary flag is set for net0")
		}
		if !test_node2.NetDevs["net1"].Primary.GetB() {
			t.Errorf("primary flag isn't set for net1")
		}
	})
	t.Run("Primary network with two networks, primary isn't set", func(t *testing.T) {
		if test_node3.PrimaryNetDev.Get() != "net0" {
			t.Errorf("primary network isn't net0 but: %s", test_node3.PrimaryNetDev.Get())
		}
		if !test_node3.NetDevs["net0"].Primary.GetB() {
			t.Errorf("primary flag is set for net0")
		}
		if test_node3.NetDevs["net1"].Primary.GetB() {
			t.Errorf("primary flag isn't set for net1")
		}
	})
	t.Run("Primary network with two networks, primary available", func(t *testing.T) {
		if test_node4.PrimaryNetDev.Get() != "net3" {
			t.Errorf("primary network isn't net3 but: %s", test_node3.PrimaryNetDev.Get())
		}
		if test_node4.NetDevs["net0"].Primary.GetB() {
			t.Errorf("primary flag is set for net0")
		}
		if test_node4.NetDevs["net1"].Primary.GetB() {
			t.Errorf("primary flag isn't set for net1")
		}
	})
}
