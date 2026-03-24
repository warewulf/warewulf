package node

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// changedSet returns a changed function that reports true for the given flag names.
func changedSet(flags ...string) func(string) bool {
	set := make(map[string]bool)
	for _, f := range flags {
		set[f] = true
	}
	return func(name string) bool { return set[name] }
}

func TestUpdateFrom_NoChanges(t *testing.T) {
	dst := NewNode("test")
	dst.Comment = "original"
	dst.ImageName = "centos7"

	src := NewNode("")
	src.Comment = "new"
	src.ImageName = "rocky9"

	dst.UpdateFrom(&src, changedSet())

	assert.Equal(t, "original", dst.Comment)
	assert.Equal(t, "centos7", dst.ImageName)
}

func TestUpdateFrom_StringField(t *testing.T) {
	dst := NewNode("test")
	dst.Comment = "original"
	dst.ImageName = "centos7"

	src := NewNode("")
	src.Comment = "updated"
	src.ImageName = "rocky9"

	dst.UpdateFrom(&src, changedSet("comment"))

	assert.Equal(t, "updated", dst.Comment)
	assert.Equal(t, "centos7", dst.ImageName, "unchanged field should be preserved")
}

func TestUpdateFrom_UndefString(t *testing.T) {
	dst := NewNode("test")
	dst.Comment = "original"

	src := NewNode("")
	src.Comment = "UNDEF"

	dst.UpdateFrom(&src, changedSet("comment"))

	assert.Equal(t, "UNDEF", dst.Comment, "UNDEF should pass through; Flatten during Persist cleans it")
}

func TestUpdateFrom_SliceField(t *testing.T) {
	dst := NewNode("test")
	dst.Profiles = []string{"default"}

	src := NewNode("")
	src.Profiles = []string{"compute", "gpu"}

	dst.UpdateFrom(&src, changedSet("profile"))

	assert.Equal(t, []string{"compute", "gpu"}, dst.Profiles)
}

func TestUpdateFrom_SliceFieldUnchanged(t *testing.T) {
	dst := NewNode("test")
	dst.Profiles = []string{"default"}

	src := NewNode("")
	src.Profiles = []string{}

	dst.UpdateFrom(&src, changedSet())

	assert.Equal(t, []string{"default"}, dst.Profiles, "unchanged slice should be preserved")
}

func TestUpdateFrom_NestedIpmiField(t *testing.T) {
	dst := NewNode("test")
	dst.Ipmi = &IpmiConf{UserName: "admin", Port: "623"}

	src := NewNode("")
	src.Ipmi = &IpmiConf{UserName: "root"}

	dst.UpdateFrom(&src, changedSet("ipmiuser"))

	assert.Equal(t, "root", dst.Ipmi.UserName)
	assert.Equal(t, "623", dst.Ipmi.Port, "unchanged ipmi field should be preserved")
}

func TestUpdateFrom_NilDstIpmiAutoCreated(t *testing.T) {
	dst := NewNode("test")
	dst.Ipmi = nil

	src := NewNode("")
	src.Ipmi = &IpmiConf{UserName: "admin"}

	dst.UpdateFrom(&src, changedSet("ipmiuser"))

	assert.NotNil(t, dst.Ipmi, "nil dst Ipmi should be auto-created")
	assert.Equal(t, "admin", dst.Ipmi.UserName)
}

func TestUpdateFrom_NetDevFieldPreservesOthers(t *testing.T) {
	dst := NewNode("test")
	dst.NetDevs = map[string]*NetDev{
		"default": {
			Ipaddr: net.ParseIP("10.0.0.1"),
			Device: "eth0",
		},
	}

	src := NewNode("")
	src.NetDevs = map[string]*NetDev{
		"default": {
			OnBoot: "true",
		},
	}

	dst.UpdateFrom(&src, changedSet("onboot"))

	assert.Equal(t, "true", string(dst.NetDevs["default"].OnBoot))
	assert.Equal(t, "10.0.0.1", dst.NetDevs["default"].Ipaddr.String(), "ipaddr should be preserved")
	assert.Equal(t, "eth0", dst.NetDevs["default"].Device, "device should be preserved")
}

func TestUpdateFrom_NetDevNewEntry(t *testing.T) {
	dst := NewNode("test")
	dst.NetDevs = map[string]*NetDev{
		"default": {Ipaddr: net.ParseIP("10.0.0.1")},
	}

	src := NewNode("")
	src.NetDevs = map[string]*NetDev{
		"secondary": {
			Device: "eth1",
		},
	}

	dst.UpdateFrom(&src, changedSet("netdev"))

	assert.Contains(t, dst.NetDevs, "default", "existing entry should be preserved")
	assert.Contains(t, dst.NetDevs, "secondary", "new entry should be created")
	assert.Equal(t, "eth1", dst.NetDevs["secondary"].Device)
	assert.Equal(t, "10.0.0.1", dst.NetDevs["default"].Ipaddr.String())
}

func TestUpdateFrom_MultipleFields(t *testing.T) {
	dst := NewNode("test")
	dst.Comment = "old"
	dst.ImageName = "centos7"
	dst.Ipmi = &IpmiConf{UserName: "admin", Port: "623"}

	src := NewNode("")
	src.Comment = "new"
	src.ImageName = "rocky9"
	src.Ipmi = &IpmiConf{UserName: "root"}

	dst.UpdateFrom(&src, changedSet("comment", "image", "ipmiuser"))

	assert.Equal(t, "new", dst.Comment)
	assert.Equal(t, "rocky9", dst.ImageName)
	assert.Equal(t, "root", dst.Ipmi.UserName)
	assert.Equal(t, "623", dst.Ipmi.Port, "unchanged ipmi field should be preserved")
}

func TestUpdateProfileFrom_BasicField(t *testing.T) {
	dst := NewProfile("test")
	dst.Comment = "original"

	src := NewProfile("")
	src.Comment = "updated"

	dst.UpdateFrom(&src, changedSet("comment"))

	assert.Equal(t, "updated", dst.Comment)
}
