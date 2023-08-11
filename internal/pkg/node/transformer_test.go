package node

import (
	"reflect"
	"strconv"
	"testing"

	"gopkg.in/yaml.v2"
)

func NewTransformerTestNode() NodeYaml {
	var data = `
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
    ipmi:
      username: greg
  profile2:
    tags:
      foo: foo profile2
    comment: Comment profile2
    ipmi:
      tags:
        foo: foo ipmi profile
nodes:
  test_node1:
    comment: Node Comment
    profiles:
    - default
    network devices:
      net0:
        device: eth1
    discoverable: true
    ipmi:
      username: chris
    tags:
      baar: baar node1
  test_node2:
    primary: net0
    profiles:
    - default
    - profile2
    network devices:
      net0:
        netmask: 1.1.1.1
      net1:
        ipaddr: 1.2.3.4
    tags:
      baar: baar node2
  test_node3:
    profiles:
    - profile2
    tags:
      foo: foo node3
      foobaar: foobaar node3
    ipmi:
      ipaddr: 1.1.1.1
      tags:
        foo: foo ipmi node3
  `
	var ret NodeYaml
	_ = yaml.Unmarshal([]byte(data), &ret)
	return ret
}
func Test_nodeYaml_SetFrom(t *testing.T) {
	c := NewTransformerTestNode()
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
	}
	getByNametests := []struct {
		name    string
		arg     string
		want    string
		wantErr bool
	}{
		{"GetByName: FieldValue", "Comment", "Node Comment", false},
		{"GetByName: FieldName", "comment", "NodeComment", true},
	}
	for _, tt := range getByNametests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetByName(&test_node1, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByName(%s,%s) error = %v, wantErr %v",
					reflect.TypeOf(test_node1), tt.arg, err, tt.wantErr)
				return
			}
			if (got != tt.want) != tt.wantErr {
				t.Errorf("GetByName(%s,%s) got = %v, want = %v",
					reflect.TypeOf(test_node1), tt.arg, got, tt.want)
				return
			}
		})
	}
	t.Run("Get() comment", func(t *testing.T) {
		comment := test_node1.Comment.Get()
		if comment != "Node Comment" {
			t.Errorf("Get() returned wrong comment: %s", comment)
		}
	})
	t.Run("Get() profile comment", func(t *testing.T) {
		comment := test_node2.Comment.Get()
		if comment != "Comment profile2" {
			t.Errorf("Get() returned wrong comment: %s", comment)
		}
	})
	t.Run("Get() default ipxe", func(t *testing.T) {
		value := test_node1.Ipxe.Get()
		if value != "default" {
			t.Errorf("Get() returned wrong ipxe template: %s", value)
		}
	})
	t.Run("GetSlice() default profile", func(t *testing.T) {
		value := test_node1.Profiles.GetSlice()[0]
		if value != "default" {
			t.Errorf("GetSlice() returned wrong profile: %s", value)
		}
	})
	t.Run("Get() default kernel args", func(t *testing.T) {
		value := test_node1.Kernel.Args.Get()
		if value != "quiet crashkernel=no vga=791 net.naming-scheme=v238" {
			t.Errorf("Get() returned wrong kernel args: %s", value)
		}
	})
	t.Run("Get() default network mask", func(t *testing.T) {
		value := test_node1.NetDevs["net0"].Netmask.Get()
		if value != "255.255.255.0" {
			t.Errorf("Get() returned wrong default netmask, got: %s want: 255.255.255.0", value)
		}
	})
	t.Run("Get() default network mask", func(t *testing.T) {
		value := test_node2.NetDevs["net0"].Netmask.Get()
		if value != "1.1.1.1" {
			t.Errorf("Get() returned wrong default netmask: %s", value)
		}
	})
	t.Run("GetB() primary for single network", func(t *testing.T) {
		value := test_node1.NetDevs["net0"].Primary.GetB()
		if !value {
			t.Errorf("GetB() returned wrong: %s", strconv.FormatBool(value))
		}
	})
	t.Run("GetB() for primary with two networks", func(t *testing.T) {
		value := test_node2.NetDevs["net0"].Primary.GetB()
		if !value {
			t.Errorf("GetB() returned wrong: %s", strconv.FormatBool(value))
		}
	})
	t.Run("GetB() for primary with two networks, get secondary network", func(t *testing.T) {
		value := test_node2.NetDevs["net1"].Primary.GetB()
		if value {
			t.Errorf("GetB() returned wrong: %s", strconv.FormatBool(value))
		}
	})
	t.Run("GetB() default discoverable", func(t *testing.T) {
		value := test_node1.Discoverable.GetB()
		if !value {
			t.Errorf("GetB() returned wrong: %s", strconv.FormatBool(value))
		}
	})
	t.Run("GetB() default discoverable", func(t *testing.T) {
		value := test_node2.Discoverable.GetB()
		if value {
			t.Errorf("GetB() returned wrong: %s", strconv.FormatBool(value))
		}
	})
	t.Run("Get() ipmi user from profile", func(t *testing.T) {
		value := test_node2.Ipmi.UserName.Get()
		if value != "greg" {
			t.Errorf("Get() returned wrong ipmi username: %s", value)
		}
	})
	t.Run("Get() ipmi user from node", func(t *testing.T) {
		value := test_node1.Ipmi.UserName.Get()
		if value != "chris" {
			t.Errorf("Get() returned wrong ipmi username: %s", value)
		}
	})
	t.Run("Get() tag foo from profile, node does not have this tag", func(t *testing.T) {
		value := test_node2.Tags["foo"].Get()
		if value != "foo profile2" {
			t.Errorf("Get() returned wrong tag for foo: %s", value)
		}
	})
	t.Run("Get() tag baar from node, node tag map is not overwritten", func(t *testing.T) {
		value := test_node2.Tags["baar"].Get()
		if value != "baar node2" {
			t.Errorf("Get() returned wrong tag for foo: %s", value)
		}
	})
	t.Run("Get() tag foo from node, tag present in profile", func(t *testing.T) {
		value := test_node3.Tags["foo"].Get()
		if value != "foo node3" {
			t.Errorf("Get() returned wrong tag for foo: %s", value)
		}
	})
	t.Run("Get() tag foobaar from node", func(t *testing.T) {
		value := test_node3.Tags["foobaar"].Get()
		if value != "foobaar node3" {
			t.Errorf("Get() returned wrong tag for foo: %s", value)
		}
	})
	t.Run("Get() ipmitag foo from profile, node does not have this tag", func(t *testing.T) {
		value := test_node3.Ipmi.Tags["foo"].Get()
		if value != "foo ipmi node3" {
			t.Errorf("Get() returned wrong tag for foo: %s", value)
		}
	})
	t.Run("Set() comment foo for empty node", func(t *testing.T) {
		test_node4.Comment.Set("foo")
		nodeConf := NewConf()
		nodeConf.GetFrom(test_node4)
		ymlByte, _ := yaml.Marshal(nodeConf)
		wanted := `comment: foo
kernel: {}
ipmi: {}
`
		if !(wanted == string(ymlByte)) {
			t.Errorf("Got wrong yml, wanted:\n'%s'\nGot:\n'%s'", wanted, string(ymlByte))
		}
		// have to remove the comment for further tests, as vscode
		// can test single functions
		test_node4.Comment.Set("UNDEF")
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ = yaml.Marshal(nodeConf)
		wanted = `{}
`
		if string(ymlByte) != wanted {
			t.Errorf("Couldn't unset comment:\n'%s'\nwanted:\n'%s'", string(ymlByte), wanted)
		}
	})

	t.Run("Set() ipmiuser foo for flattened empty node", func(t *testing.T) {
		test_node4.Ipmi.UserName.Set("foo")
		nodeConf := NewConf()
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ := yaml.Marshal(nodeConf)
		wanted := `ipmi:
  username: foo
`
		if !(wanted == string(ymlByte)) {
			t.Errorf("Got wrong yml, wanted:\n'%s'\nGot:\n'%s'", wanted, string(ymlByte))
		}
		test_node4.Ipmi.Tags["foo"] = &Entry{}
		test_node4.Ipmi.Tags["foo"].Set("baar")
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ = yaml.Marshal(nodeConf)
		wanted = `ipmi:
  username: foo
  tags:
    foo: baar
`
		if !(wanted == string(ymlByte)) {
			t.Errorf("Got wrong yml, wanted:\n'%s'\nGot:\n'%s'", wanted, string(ymlByte))
		}
		test_node4.Ipmi.UserName.Set("UNSET")
		delete(test_node4.Ipmi.Tags, "foo")
	})
	t.Run("Set() kernelargs foo for flattened empty node", func(t *testing.T) {
		test_node4.Kernel.Args.Set("foo")
		nodeConf := NewConf()
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ := yaml.Marshal(nodeConf)
		wanted := `kernel:
  args: foo
`
		if !(wanted == string(ymlByte)) {
			t.Errorf("Got wrong yml, wanted:\n'%s'\nGot:\n'%s'", wanted, string(ymlByte))
		}
		test_node4.Kernel.Args.Set("--")
	})
	t.Run("Set() tag foo to bar for flattened empty node", func(t *testing.T) {
		test_node4.Tags["foo"] = &Entry{}
		test_node4.Tags["foo"].Set("baar")
		nodeConf := NewConf()
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ := yaml.Marshal(nodeConf)
		wanted := `tags:
  foo: baar
`
		if !(wanted == string(ymlByte)) {
			t.Errorf("Got wrong yml, wanted:\n'%s'\nGot:\n'%s'", wanted, string(ymlByte))
		}
		delete(test_node4.Tags, "foo")
		nodeConf = NewConf()
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ = yaml.Marshal(nodeConf)
		wanted = `{}
`
		if string(ymlByte) != wanted {
			t.Errorf("Couldn't remove tag, wanted:\n%s\nGot:\n%s", wanted, string(ymlByte))
		}

	})
	t.Run("Set() netdev foo with device name baar for flattened empty node", func(t *testing.T) {
		test_node4.NetDevs["foo"] = new(NetDevEntry)
		test_node4.NetDevs["foo"].Device.Set("baar")
		nodeConf := NewConf()
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ := yaml.Marshal(nodeConf)
		wanted := `network devices:
  foo:
    device: baar
`
		if !(wanted == string(ymlByte)) {
			t.Errorf("Got wrong yml, wanted:\n'%s'\nGot:\n'%s'", wanted, string(ymlByte))
		}
		test_node4.NetDevs["foo"].Tags = make(map[string]*Entry)
		test_node4.NetDevs["foo"].Tags["netfoo"] = new(Entry)
		test_node4.NetDevs["foo"].Tags["netfoo"].Set("netbaar")
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		wanted = `network devices:
  foo:
    device: baar
    tags:
      netfoo: netbaar
`
		ymlByte, _ = yaml.Marshal(nodeConf)
		if string(ymlByte) != wanted {
			t.Errorf("Couldn't set nettag: '%s' got: '%s'", wanted, string(ymlByte))
		}

		delete(test_node4.NetDevs, "foo")
		nodeConf = NewConf()
		nodeConf.GetFrom(test_node4)
		nodeConf.Flatten()
		ymlByte, _ = yaml.Marshal(nodeConf)
		wanted = `{}
`
		if string(ymlByte) != wanted {
			t.Errorf("Couldn't remove tag'%s'", string(ymlByte))
		}
	})
}
