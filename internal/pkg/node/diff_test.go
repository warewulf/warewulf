package node

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffProfile_NoChange(t *testing.T) {
	a := NewNode("n01")
	a.Comment = "hello"
	a.SystemOverlay = []string{"wwinit"}
	b := a.Clone()

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Empty(t, changes)
}

func TestDiffProfile_ScalarString(t *testing.T) {
	a := NewNode("n01")
	a.Comment = "old"
	b := a.Clone()
	b.Comment = "new"

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Equal(t, []Change{{Path: "comment", Before: `"old"`, After: `"new"`}}, changes)
}

func TestDiffProfile_ScalarFromUnset(t *testing.T) {
	a := NewNode("n01")
	b := a.Clone()
	b.Comment = "hello"

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Equal(t, []Change{{Path: "comment", Before: "<unset>", After: `"hello"`}}, changes)
}

func TestDiffProfile_Slice(t *testing.T) {
	a := NewNode("n01")
	a.SystemOverlay = []string{"wwinit", "wwclient"}
	b := a.Clone()
	b.SystemOverlay = []string{"wwinit", "wwclient", "foo"}

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Equal(t, []Change{{
		Path:   "system-overlays",
		Before: "[wwinit, wwclient]",
		After:  "[wwinit, wwclient, foo]",
	}}, changes)
}

func TestDiffProfile_NestedKernelArgs(t *testing.T) {
	a := NewNode("n01")
	a.Kernel = &KernelConf{Args: []string{"quiet"}}
	b := a.Clone()
	b.Kernel.Args = []string{"quiet", "console=ttyS0"}

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Equal(t, []Change{{
		Path:   "kernel.kernelargs",
		Before: "[quiet]",
		After:  "[quiet, console=ttyS0]",
	}}, changes)
}

func TestDiffProfile_PointerAutoCreated(t *testing.T) {
	a := NewNode("n01")
	b := a.Clone()
	b.Ipmi = &IpmiConf{UserName: "root"}

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Contains(t, changes, Change{Path: "ipmi.ipmiuser", Before: "<unset>", After: `"root"`})
}

func TestDiffProfile_TagMap(t *testing.T) {
	a := NewNode("n01")
	a.Tags = map[string]string{"role": "compute"}
	b := a.Clone()
	b.Tags = map[string]string{"role": "gpu", "rack": "A1"}

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Contains(t, changes, Change{Path: "tags[role]", Before: `"compute"`, After: `"gpu"`})
	assert.Contains(t, changes, Change{Path: "tags[rack]", Before: "<unset>", After: `"A1"`})
}

func TestDiffProfile_NetDevField(t *testing.T) {
	a := NewNode("n01")
	a.NetDevs = map[string]*NetDev{
		"default": {Ipaddr: net.ParseIP("10.0.0.1"), Device: "eth0"},
	}
	b := a.Clone()
	b.NetDevs["default"].Ipaddr = net.ParseIP("10.0.0.2")

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Contains(t, changes, Change{Path: "netdevs[default].ipaddr", Before: "10.0.0.1", After: "10.0.0.2"})
}

func TestFormatChanges_Collapse(t *testing.T) {
	common := []Change{{Path: "system-overlays", Before: "[wwinit]", After: "[wwinit, foo]"}}
	uniq := []Change{{Path: "comment", Before: "<unset>", After: `"x"`}}

	out := FormatChanges(map[string][]Change{
		"n01": common,
		"n02": common,
		"n03": uniq,
	})

	assert.Equal(t, "n01, n02:\n  system-overlays: [wwinit] → [wwinit, foo]\n\nn03:\n  comment: <unset> → \"x\"\n", out)
}

func TestFormatChanges_Empty(t *testing.T) {
	assert.Equal(t, "", FormatChanges(map[string][]Change{}))
	assert.Equal(t, "", FormatChanges(map[string][]Change{"n01": nil}))
}

func TestDiffProfile_NetDevAdded(t *testing.T) {
	a := NewNode("n01")
	a.NetDevs = map[string]*NetDev{
		"default": {Device: "eth0"},
	}
	b := a.Clone()
	b.NetDevs["secondary"] = &NetDev{Device: "eth1"}

	changes := DiffProfile(&a.Profile, &b.Profile)
	assert.Contains(t, changes, Change{Path: "netdevs[secondary].netdev", Before: "<unset>", After: `"eth1"`})
}
