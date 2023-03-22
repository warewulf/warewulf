package node


import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)


func writeTestConfigFile(data string) (f *os.File, err error) {
	f, err = ioutil.TempFile("", "nodes.conf-*")
	if err != nil {
		return f, err
	} else {
		_, err = f.WriteString(data)
		if err != nil {
			return f, err
		} else {
			err = f.Sync()
			return f, err
		}
	}
}


func Test_ReadNodeYamlFromFileMinimal(t *testing.T) {
	file, writeErr := writeTestConfigFile(`
nodeprofiles:
  default:
    comment: A default profile
nodes:
  test_node:
    comment: A single node`)
	if file != nil {
		defer os.Remove(file.Name())
	}
	assert.NoError(t, writeErr)

	nodeYaml, err := ReadNodeYamlFromFile(file.Name())
	assert.NoError(t, err)
	assert.Contains(t, nodeYaml.NodeProfiles, "default")
	assert.Equal(t, "A default profile", nodeYaml.NodeProfiles["default"].Comment)
	assert.Contains(t, nodeYaml.Nodes, "test_node")
	assert.Equal(t, "A single node", nodeYaml.Nodes["test_node"].Comment)
}


func Test_GetAllNodeInfoDefaults(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodes:
  test_node: {}`))
	assert.NoError(t, err)
	allNodeInfo, _ := nodeYaml.GetAllNodeInfo()
	nodeInfo := allNodeInfo[0]
	assert.Equal(t, "test_node", nodeInfo.Id.Get())
	assert.Empty(t, nodeInfo.Comment.Get())
	assert.Empty(t, nodeInfo.ClusterName.Get())
	assert.Empty(t, nodeInfo.ContainerName.Get())
	assert.Equal(t, "default", nodeInfo.Ipxe.Get())
	assert.Len(t, nodeInfo.RuntimeOverlay.GetSlice(), 1)
	assert.Contains(t, nodeInfo.RuntimeOverlay.GetSlice(), "generic")
	assert.Len(t, nodeInfo.SystemOverlay.GetSlice(), 1)
	assert.Contains(t, nodeInfo.SystemOverlay.GetSlice(), "wwinit")

	assert.Empty(t, nodeInfo.Kernel.Override.Get())
	assert.Equal(t, "quiet crashkernel=no vga=791 net.naming-scheme=v238", nodeInfo.Kernel.Args.Get())

	assert.Empty(t, nodeInfo.Ipmi.Ipaddr.Get())
	assert.Empty(t, nodeInfo.Ipmi.Netmask.Get())
	assert.Empty(t, nodeInfo.Ipmi.Port.Get())
	assert.Empty(t, nodeInfo.Ipmi.Gateway.Get())
	assert.Empty(t, nodeInfo.Ipmi.UserName.Get())
	assert.Empty(t, nodeInfo.Ipmi.Password.Get())
	assert.Empty(t, nodeInfo.Ipmi.Interface.Get())
	assert.False(t, nodeInfo.Ipmi.Write.GetB())
	assert.Empty(t, nodeInfo.Ipmi.Tags)

	assert.Equal(t, "/sbin/init", nodeInfo.Init.Get())
	assert.Equal(t, "initramfs", nodeInfo.Root.Get())
	assert.Empty(t, nodeInfo.AssetKey.Get())
	assert.False(t, nodeInfo.Discoverable.GetB())
	assert.Len(t, nodeInfo.Profiles.GetSlice(), 1)
	assert.Contains(t, nodeInfo.Profiles.GetSlice(), "default")
	assert.Empty(t, nodeInfo.NetDevs)
	assert.Empty(t, nodeInfo.Tags)
	assert.Empty(t, nodeInfo.PrimaryNetDev.Get())
}


func Test_GetAllNodeInfoFull(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodes:
  test_node.cluster.example.net:
    comment: A single node
    container name: linux-compute
    ipxe template: local-ipxe-template
    runtime overlay:
    - runtime1
    - runtime2
    system overlay:
    - system1
    - system2
    kernel:
      override: v0.0.1-rc1
      args: init=/bin/sh
    ipmi:
      ipaddr: 127.0.0.1
      netmask: 255.255.255.0
      port: 163
      gateway: 127.0.0.254
      username: root
      password: calvin
      interface: ipmi0
      write: true
      tags:
        IPMITagKey: IPMITagValue
    init: /sbin/systemd
    root: tmpfs
    asset key: ASDF123
    discoverable: true
    profiles:
    - compute
    network devices:
      fabric:
        type: infiniband
        onboot: true
        device: ib0
        hwaddr: 00:FF:00:FF:00:FF
        ipaddr: 127.0.0.2
        ip6addr: "::2"
        ipcidr: 127.0.0.2/24
        prefix: 24
        netmask: 255.255.255.0
        gateway: 127.0.0.254
        mtu: 9000
        tags:
          ib0TagKey: ib0TagValue
    primary network: fabric
    tags:
      TestTagKey: TestTagValue`))
	assert.NoError(t, err)
	allNodeInfo, _ := nodeYaml.GetAllNodeInfo()
	nodeInfo := allNodeInfo[0]
	assert.Equal(t, "test_node.cluster.example.net", nodeInfo.Id.Get())
	assert.Equal(t, "A single node", nodeInfo.Comment.Get())
	assert.Equal(t, "cluster.example.net", nodeInfo.ClusterName.Get())
	assert.Equal(t, "linux-compute", nodeInfo.ContainerName.Get())
	assert.Equal(t, "local-ipxe-template", nodeInfo.Ipxe.Get())
	assert.Equal(t, []string{"runtime1", "runtime2"}, nodeInfo.RuntimeOverlay.GetSlice())
	assert.Equal(t, []string{"system1", "system2"}, nodeInfo.SystemOverlay.GetSlice())
	assert.Equal(t, "v0.0.1-rc1", nodeInfo.Kernel.Override.Get())
	assert.Equal(t, "init=/bin/sh", nodeInfo.Kernel.Args.Get())

	assert.Equal(t, "127.0.0.1", nodeInfo.Ipmi.Ipaddr.Get())
	assert.Equal(t, "255.255.255.0", nodeInfo.Ipmi.Netmask.Get())
	assert.Equal(t, "163", nodeInfo.Ipmi.Port.Get())
	assert.Equal(t, "127.0.0.254", nodeInfo.Ipmi.Gateway.Get())
	assert.Equal(t, "root", nodeInfo.Ipmi.UserName.Get())
	assert.Equal(t, "calvin", nodeInfo.Ipmi.Password.Get())
	assert.Equal(t, "ipmi0", nodeInfo.Ipmi.Interface.Get())
	assert.True(t, nodeInfo.Ipmi.Write.GetB())
	assert.Len(t, nodeInfo.Ipmi.Tags, 1)
	assert.Equal(t, "IPMITagValue", nodeInfo.Ipmi.Tags["IPMITagKey"].Get())
	
	assert.Equal(t, "/sbin/systemd", nodeInfo.Init.Get())
	assert.Equal(t, "tmpfs", nodeInfo.Root.Get())
	assert.Equal(t, "ASDF123", nodeInfo.AssetKey.Get())
	assert.True(t, nodeInfo.Discoverable.GetB())
	assert.Equal(t, []string{"compute"}, nodeInfo.Profiles.GetSlice(), 1)
	assert.Len(t, nodeInfo.NetDevs, 1)
	assert.Contains(t, nodeInfo.NetDevs, "fabric")
	assert.Equal(t, "infiniband", nodeInfo.NetDevs["fabric"].Type.Get())
	assert.True(t, nodeInfo.NetDevs["fabric"].OnBoot.GetB())
	assert.Equal(t, "ib0", nodeInfo.NetDevs["fabric"].Device.Get())
	assert.Equal(t, "00:FF:00:FF:00:FF", nodeInfo.NetDevs["fabric"].Hwaddr.Get())
	assert.Equal(t, "127.0.0.2", nodeInfo.NetDevs["fabric"].Ipaddr.Get())
	assert.Equal(t, "::2", nodeInfo.NetDevs["fabric"].Ipaddr6.Get())
	assert.Equal(t, "127.0.0.2/24", nodeInfo.NetDevs["fabric"].IpCIDR.Get())
	assert.Equal(t, "24", nodeInfo.NetDevs["fabric"].Prefix.Get())
	assert.Equal(t, "255.255.255.0", nodeInfo.NetDevs["fabric"].Netmask.Get())
	assert.Equal(t, "127.0.0.254", nodeInfo.NetDevs["fabric"].Gateway.Get())
	assert.Equal(t, "9000", nodeInfo.NetDevs["fabric"].MTU.Get())
	assert.True(t, nodeInfo.NetDevs["fabric"].Primary.GetB())
	assert.Len(t, nodeInfo.NetDevs["fabric"].Tags, 1)
	assert.Contains(t, nodeInfo.NetDevs["fabric"].Tags, "ib0TagKey")
	assert.Equal(t, "ib0TagValue", nodeInfo.NetDevs["fabric"].Tags["ib0TagKey"].Get())
	assert.Equal(t, "fabric", nodeInfo.PrimaryNetDev.Get())
	assert.Len(t, nodeInfo.Tags, 1)
	assert.Contains(t, nodeInfo.Tags, "TestTagKey")
	assert.Equal(t, "TestTagValue", nodeInfo.Tags["TestTagKey"].Get())
}

func Test_GetAllNodeInfoDefaultNetDev(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodes:
  test_node:
    network devices:
      default: {}`))
	assert.NoError(t, err)
	allNodeInfo, _ := nodeYaml.GetAllNodeInfo()
	nodeInfo := allNodeInfo[0]
	assert.Equal(t, "eth0", nodeInfo.NetDevs["default"].Device.Get())
	assert.Equal(t, "ethernet", nodeInfo.NetDevs["default"].Type.Get())
	assert.Equal(t, "255.255.255.0", nodeInfo.NetDevs["default"].Netmask.Get())
	assert.True(t, nodeInfo.NetDevs["default"].Primary.GetB())
}

func Test_GetAllNodeInfoCompatibility(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodes:
  test_node.cluster.example.net:
    kernel version: v0.0.1
    kernel override: v0.0.1-rc2
    kernel args: init=/bin/sh

    ipmi ipaddr: 127.0.0.1
    ipmi netmask: 255.255.255.0
    ipmi port: 163
    ipmi gateway: 127.0.0.254
    ipmi username: root
    ipmi password: calvin
    ipmi interface: ipmi0
    ipmi write: true

    keys:
      TestKeyKey: TestKeyValue`))
	assert.NoError(t, err)
	allNodeInfo, _ := nodeYaml.GetAllNodeInfo()
	nodeInfo := allNodeInfo[0]
	
	assert.Equal(t, "v0.0.1-rc2", nodeInfo.Kernel.Override.Get())
	assert.Equal(t, "init=/bin/sh", nodeInfo.Kernel.Args.Get())

	assert.Equal(t, "127.0.0.1", nodeInfo.Ipmi.Ipaddr.Get())
	assert.Equal(t, "255.255.255.0", nodeInfo.Ipmi.Netmask.Get())
	assert.Equal(t, "163", nodeInfo.Ipmi.Port.Get())
	assert.Equal(t, "127.0.0.254", nodeInfo.Ipmi.Gateway.Get())
	assert.Equal(t, "root", nodeInfo.Ipmi.UserName.Get())
	assert.Equal(t, "calvin", nodeInfo.Ipmi.Password.Get())
	assert.Equal(t, "ipmi0", nodeInfo.Ipmi.Interface.Get())
	assert.True(t, nodeInfo.Ipmi.Write.GetB())

	assert.Equal(t, "TestKeyValue", nodeInfo.Tags["TestKeyKey"].Get())
}


func Test_GetAllNodeInfoCompatibilityPrecedence(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodes:
  test_node.cluster.example.net:
    kernel version: v0.0.1
    kernel override: v0.0.1-rc2
    kernel args: init=/bin/sh

    ipmi ipaddr: 127.0.0.1
    ipmi netmask: 255.255.255.0
    ipmi port: 163
    ipmi gateway: 127.0.0.254
    ipmi username: root
    ipmi password: calvin
    ipmi interface: ipmi0
    ipmi write: true

    keys:
      TestKeyKey: TestKeyValue

    kernel:
      override: v0.2.1-rc1
      args: init=/bin/bash

    ipmi:
      ipaddr: 127.0.2.1
      netmask: 255.255.0.0
      port: 8163
      gateway: 127.0.2.254
      username: admin
      password: hobbes
      interface: ipmi1
      write: false

    tags:
      TestTagKey: TestTagValue`))
	assert.NoError(t, err)
	allNodeInfo, _ := nodeYaml.GetAllNodeInfo()
	nodeInfo := allNodeInfo[0]
	
	assert.Equal(t, "v0.2.1-rc1", nodeInfo.Kernel.Override.Get())
	assert.Equal(t, "init=/bin/bash", nodeInfo.Kernel.Args.Get())

	assert.Equal(t, "127.0.2.1", nodeInfo.Ipmi.Ipaddr.Get())
	assert.Equal(t, "255.255.0.0", nodeInfo.Ipmi.Netmask.Get())
	assert.Equal(t, "8163", nodeInfo.Ipmi.Port.Get())
	assert.Equal(t, "127.0.2.254", nodeInfo.Ipmi.Gateway.Get())
	assert.Equal(t, "admin", nodeInfo.Ipmi.UserName.Get())
	assert.Equal(t, "hobbes", nodeInfo.Ipmi.Password.Get())
	assert.Equal(t, "ipmi1", nodeInfo.Ipmi.Interface.Get())
	assert.False(t, nodeInfo.Ipmi.Write.GetB())

	assert.Len(t, nodeInfo.Tags, 2)
	assert.Contains(t, nodeInfo.Tags, "TestKeyKey")
	assert.Equal(t, "TestKeyValue", nodeInfo.Tags["TestKeyKey"].Get())
	assert.Contains(t, nodeInfo.Tags, "TestTagKey")
	assert.Equal(t, "TestTagValue", nodeInfo.Tags["TestTagKey"].Get())
}


func Test_GetAllNodeInfoFullProfile(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodes:
  test_node.cluster.example.net:
    profiles:
    - test_profile
nodeprofiles:
  test_profile:
    comment: A single node
    container name: linux-compute
    ipxe template: local-ipxe-template
    runtime overlay:
    - runtime1
    - runtime2
    system overlay:
    - system1
    - system2
    kernel:
      override: v0.0.1-rc1
      args: init=/bin/sh
    ipmi:
      ipaddr: 127.0.0.1
      netmask: 255.255.255.0
      port: 163
      gateway: 127.0.0.254
      username: root
      password: calvin
      interface: ipmi0
      write: true
      tags:
        IPMITagKey: IPMITagValue
    init: /sbin/systemd
    root: tmpfs
    asset key: ASDF123
    discoverable: true
    network devices:
      fabric:
        type: infiniband
        onboot: true
        device: ib0
        hwaddr: 00:FF:00:FF:00:FF
        ipaddr: 127.0.0.2
        ip6addr: "::2"
        ipcidr: 127.0.0.2/24
        prefix: 24
        netmask: 255.255.255.0
        gateway: 127.0.0.254
        mtu: 9000
        tags:
          ib0TagKey: ib0TagValue
    primary network: fabric
    tags:
      TestTagKey: TestTagValue`))
	assert.NoError(t, err)
	allNodeInfo, _ := nodeYaml.GetAllNodeInfo()
	nodeInfo := allNodeInfo[0]
	assert.Equal(t, "test_node.cluster.example.net", nodeInfo.Id.Get())
	assert.Equal(t, "A single node", nodeInfo.Comment.Get())
	assert.Equal(t, "cluster.example.net", nodeInfo.ClusterName.Get())
	assert.Equal(t, "linux-compute", nodeInfo.ContainerName.Get())
	assert.Equal(t, "local-ipxe-template", nodeInfo.Ipxe.Get())
	assert.Equal(t, []string{"runtime1", "runtime2"}, nodeInfo.RuntimeOverlay.GetSlice())
	assert.Equal(t, []string{"system1", "system2"}, nodeInfo.SystemOverlay.GetSlice())
	assert.Equal(t, "v0.0.1-rc1", nodeInfo.Kernel.Override.Get())
	assert.Equal(t, "init=/bin/sh", nodeInfo.Kernel.Args.Get())

	assert.Equal(t, "127.0.0.1", nodeInfo.Ipmi.Ipaddr.Get())
	assert.Equal(t, "255.255.255.0", nodeInfo.Ipmi.Netmask.Get())
	assert.Equal(t, "163", nodeInfo.Ipmi.Port.Get())
	assert.Equal(t, "127.0.0.254", nodeInfo.Ipmi.Gateway.Get())
	assert.Equal(t, "root", nodeInfo.Ipmi.UserName.Get())
	assert.Equal(t, "calvin", nodeInfo.Ipmi.Password.Get())
	assert.Equal(t, "ipmi0", nodeInfo.Ipmi.Interface.Get())
	assert.True(t, nodeInfo.Ipmi.Write.GetB())
	assert.Len(t, nodeInfo.Ipmi.Tags, 1)
	assert.Equal(t, "IPMITagValue", nodeInfo.Ipmi.Tags["IPMITagKey"].Get())
	
	assert.Equal(t, "/sbin/systemd", nodeInfo.Init.Get())
	assert.Equal(t, "tmpfs", nodeInfo.Root.Get())
	assert.Equal(t, "ASDF123", nodeInfo.AssetKey.Get())
	assert.True(t, nodeInfo.Discoverable.GetB())
	assert.Equal(t, []string{"test_profile"}, nodeInfo.Profiles.GetSlice(), 1)
	assert.Len(t, nodeInfo.NetDevs, 1)
	assert.Contains(t, nodeInfo.NetDevs, "fabric")
	assert.Equal(t, "infiniband", nodeInfo.NetDevs["fabric"].Type.Get())
	assert.True(t, nodeInfo.NetDevs["fabric"].OnBoot.GetB())
	assert.Equal(t, "ib0", nodeInfo.NetDevs["fabric"].Device.Get())
	assert.Equal(t, "00:FF:00:FF:00:FF", nodeInfo.NetDevs["fabric"].Hwaddr.Get())
	assert.Equal(t, "127.0.0.2", nodeInfo.NetDevs["fabric"].Ipaddr.Get())
	assert.Equal(t, "::2", nodeInfo.NetDevs["fabric"].Ipaddr6.Get())
	assert.Equal(t, "127.0.0.2/24", nodeInfo.NetDevs["fabric"].IpCIDR.Get())
	assert.Equal(t, "24", nodeInfo.NetDevs["fabric"].Prefix.Get())
	assert.Equal(t, "255.255.255.0", nodeInfo.NetDevs["fabric"].Netmask.Get())
	assert.Equal(t, "127.0.0.254", nodeInfo.NetDevs["fabric"].Gateway.Get())
	assert.Equal(t, "9000", nodeInfo.NetDevs["fabric"].MTU.Get())
	assert.True(t, nodeInfo.NetDevs["fabric"].Primary.GetB())
	assert.Len(t, nodeInfo.NetDevs["fabric"].Tags, 1)
	assert.Contains(t, nodeInfo.NetDevs["fabric"].Tags, "ib0TagKey")
	assert.Equal(t, "ib0TagValue", nodeInfo.NetDevs["fabric"].Tags["ib0TagKey"].Get())
	assert.Equal(t, "fabric", nodeInfo.PrimaryNetDev.Get())
	assert.Len(t, nodeInfo.Tags, 1)
	assert.Contains(t, nodeInfo.Tags, "TestTagKey")
	assert.Equal(t, "TestTagValue", nodeInfo.Tags["TestTagKey"].Get())
}


func Test_GetAllProfileInfoEmpty(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(""))
	assert.NoError(t, err)
	allProfileInfo, _ := nodeYaml.GetAllProfileInfo()
	assert.Len(t, allProfileInfo, 0)
}


func Test_GetAllProfileInfoFull(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodeprofiles:
  test_profile.cluster.example.net:
    comment: A single node
    container name: linux-compute
    ipxe template: local-ipxe-template
    runtime overlay:
    - runtime1
    - runtime2
    system overlay:
    - system1
    - system2
    kernel:
      override: v0.0.1-rc1
      args: init=/bin/sh
    ipmi:
      ipaddr: 127.0.0.1
      netmask: 255.255.255.0
      port: 163
      gateway: 127.0.0.254
      username: root
      password: calvin
      interface: ipmi0
      write: true
      tags:
        IPMITagKey: IPMITagValue
    init: /sbin/systemd
    root: tmpfs
    asset key: ASDF123
    discoverable: true
    network devices:
      fabric:
        type: infiniband
        onboot: true
        device: ib0
        hwaddr: 00:FF:00:FF:00:FF
        ipaddr: 127.0.0.2
        ip6addr: "::2"
        ipcidr: 127.0.0.2/24
        prefix: 24
        netmask: 255.255.255.0
        gateway: 127.0.0.254
        mtu: 9000
        tags:
          ib0TagKey: ib0TagValue
    primary network: fabric
    tags:
      TestTagKey: TestTagValue`))
	assert.NoError(t, err)
	allProfileInfo, _ := nodeYaml.GetAllProfileInfo()
	profileInfo := allProfileInfo[0]
	assert.Equal(t, "test_profile.cluster.example.net", profileInfo.Id.Get())
	assert.Equal(t, "A single node", profileInfo.Comment.Get())
	assert.Equal(t, "cluster.example.net", profileInfo.ClusterName.Get())
	assert.Equal(t, "linux-compute", profileInfo.ContainerName.Get())
	assert.Equal(t, "local-ipxe-template", profileInfo.Ipxe.Get())
	assert.Equal(t, []string{"runtime1", "runtime2"}, profileInfo.RuntimeOverlay.GetSlice())
	assert.Equal(t, []string{"system1", "system2"}, profileInfo.SystemOverlay.GetSlice())
	assert.Equal(t, "v0.0.1-rc1", profileInfo.Kernel.Override.Get())
	assert.Equal(t, "init=/bin/sh", profileInfo.Kernel.Args.Get())

	assert.Equal(t, "127.0.0.1", profileInfo.Ipmi.Ipaddr.Get())
	assert.Equal(t, "255.255.255.0", profileInfo.Ipmi.Netmask.Get())
	assert.Equal(t, "163", profileInfo.Ipmi.Port.Get())
	assert.Equal(t, "127.0.0.254", profileInfo.Ipmi.Gateway.Get())
	assert.Equal(t, "root", profileInfo.Ipmi.UserName.Get())
	assert.Equal(t, "calvin", profileInfo.Ipmi.Password.Get())
	assert.Equal(t, "ipmi0", profileInfo.Ipmi.Interface.Get())
	assert.True(t, profileInfo.Ipmi.Write.GetB())
	assert.Len(t, profileInfo.Ipmi.Tags, 1)
	assert.Equal(t, "IPMITagValue", profileInfo.Ipmi.Tags["IPMITagKey"].Get())
	
	assert.Equal(t, "/sbin/systemd", profileInfo.Init.Get())
	assert.Equal(t, "tmpfs", profileInfo.Root.Get())
	assert.Equal(t, "ASDF123", profileInfo.AssetKey.Get())
	assert.True(t, profileInfo.Discoverable.GetB())
	assert.Len(t, profileInfo.NetDevs, 1)
	assert.Contains(t, profileInfo.NetDevs, "fabric")
	assert.Equal(t, "infiniband", profileInfo.NetDevs["fabric"].Type.Get())
	assert.True(t, profileInfo.NetDevs["fabric"].OnBoot.GetB())
	assert.Equal(t, "ib0", profileInfo.NetDevs["fabric"].Device.Get())
	assert.Equal(t, "00:FF:00:FF:00:FF", profileInfo.NetDevs["fabric"].Hwaddr.Get())
	assert.Equal(t, "127.0.0.2", profileInfo.NetDevs["fabric"].Ipaddr.Get())
	assert.Equal(t, "::2", profileInfo.NetDevs["fabric"].Ipaddr6.Get())
	assert.Equal(t, "127.0.0.2/24", profileInfo.NetDevs["fabric"].IpCIDR.Get())
	assert.Equal(t, "24", profileInfo.NetDevs["fabric"].Prefix.Get())
	assert.Equal(t, "255.255.255.0", profileInfo.NetDevs["fabric"].Netmask.Get())
	assert.Equal(t, "127.0.0.254", profileInfo.NetDevs["fabric"].Gateway.Get())
	assert.Equal(t, "9000", profileInfo.NetDevs["fabric"].MTU.Get())
	assert.True(t, profileInfo.NetDevs["fabric"].Primary.GetB())
	assert.Len(t, profileInfo.NetDevs["fabric"].Tags, 1)
	assert.Contains(t, profileInfo.NetDevs["fabric"].Tags, "ib0TagKey")
	assert.Equal(t, "ib0TagValue", profileInfo.NetDevs["fabric"].Tags["ib0TagKey"].Get())
	assert.Equal(t, "fabric", profileInfo.PrimaryNetDev.Get())
	assert.Len(t, profileInfo.Tags, 1)
	assert.Contains(t, profileInfo.Tags, "TestTagKey")
	assert.Equal(t, "TestTagValue", profileInfo.Tags["TestTagKey"].Get())
}


func Test_GetAllProfileInfoCompatibility(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodeprofiles:
  test_profile.cluster.example.net:
    kernel version: v0.0.1
    kernel override: v0.0.1-rc2
    kernel args: init=/bin/sh

    ipmi ipaddr: 127.0.0.1
    ipmi netmask: 255.255.255.0
    ipmi port: 163
    ipmi gateway: 127.0.0.254
    ipmi username: root
    ipmi password: calvin
    ipmi interface: ipmi0
    ipmi write: true

    keys:
      TestKeyKey: TestKeyValue`))
	assert.NoError(t, err)
	allProfileInfo, _ := nodeYaml.GetAllProfileInfo()
	profileInfo := allProfileInfo[0]
	
	assert.Equal(t, "v0.0.1-rc2", profileInfo.Kernel.Override.Get())
	assert.Equal(t, "init=/bin/sh", profileInfo.Kernel.Args.Get())

	assert.Equal(t, "127.0.0.1", profileInfo.Ipmi.Ipaddr.Get())
	assert.Equal(t, "255.255.255.0", profileInfo.Ipmi.Netmask.Get())
	assert.Equal(t, "163", profileInfo.Ipmi.Port.Get())
	assert.Equal(t, "127.0.0.254", profileInfo.Ipmi.Gateway.Get())
	assert.Equal(t, "root", profileInfo.Ipmi.UserName.Get())
	assert.Equal(t, "calvin", profileInfo.Ipmi.Password.Get())
	assert.Equal(t, "ipmi0", profileInfo.Ipmi.Interface.Get())
	assert.True(t, profileInfo.Ipmi.Write.GetB())

	assert.Equal(t, "TestKeyValue", profileInfo.Tags["TestKeyKey"].Get())
}


func Test_GetAllProfileInfoCompatibilityPrecedence(t *testing.T) {
	nodeYaml, err := ParseNodeYaml([]byte(`
nodeprofiles:
  test_node.cluster.example.net:
    kernel version: v0.0.1
    kernel override: v0.0.1-rc2
    kernel args: init=/bin/sh

    ipmi ipaddr: 127.0.0.1
    ipmi netmask: 255.255.255.0
    ipmi port: 163
    ipmi gateway: 127.0.0.254
    ipmi username: root
    ipmi password: calvin
    ipmi interface: ipmi0
    ipmi write: true

    keys:
      TestKeyKey: TestKeyValue

    kernel:
      override: v0.2.1-rc1
      args: init=/bin/bash

    ipmi:
      ipaddr: 127.0.2.1
      netmask: 255.255.0.0
      port: 8163
      gateway: 127.0.2.254
      username: admin
      password: hobbes
      interface: ipmi1
      write: false

    tags:
      TestTagKey: TestTagValue`))
	assert.NoError(t, err)
	allProfileInfo, _ := nodeYaml.GetAllProfileInfo()
	profileInfo := allProfileInfo[0]
	
	assert.Equal(t, "v0.2.1-rc1", profileInfo.Kernel.Override.Get())
	assert.Equal(t, "init=/bin/bash", profileInfo.Kernel.Args.Get())

	assert.Equal(t, "127.0.2.1", profileInfo.Ipmi.Ipaddr.Get())
	assert.Equal(t, "255.255.0.0", profileInfo.Ipmi.Netmask.Get())
	assert.Equal(t, "8163", profileInfo.Ipmi.Port.Get())
	assert.Equal(t, "127.0.2.254", profileInfo.Ipmi.Gateway.Get())
	assert.Equal(t, "admin", profileInfo.Ipmi.UserName.Get())
	assert.Equal(t, "hobbes", profileInfo.Ipmi.Password.Get())
	assert.Equal(t, "ipmi1", profileInfo.Ipmi.Interface.Get())
	assert.False(t, profileInfo.Ipmi.Write.GetB())

	assert.Len(t, profileInfo.Tags, 2)
	assert.Contains(t, profileInfo.Tags, "TestKeyKey")
	assert.Equal(t, "TestKeyValue", profileInfo.Tags["TestKeyKey"].Get())
	assert.Contains(t, profileInfo.Tags, "TestTagKey")
	assert.Equal(t, "TestTagValue", profileInfo.Tags["TestTagKey"].Get())
}


// Test sorting of nodes
// Test sorting of profiles
