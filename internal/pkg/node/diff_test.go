package node

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff_NoChange(t *testing.T) {
	a := NewNode("n01")
	a.Comment = "hello"
	a.SystemOverlay = []string{"wwinit"}
	b := a.Clone()

	changes := Diff(&a.Profile, &b.Profile)
	assert.Empty(t, changes)
}

func TestDiff_ScalarString(t *testing.T) {
	a := NewNode("n01")
	a.Comment = "old"
	b := a.Clone()
	b.Comment = "new"

	changes := Diff(&a.Profile, &b.Profile)
	assert.Equal(t, []Change{{Path: "comment", Before: `"old"`, After: `"new"`}}, changes)
}

func TestDiff_Slice(t *testing.T) {
	a := NewNode("n01")
	a.SystemOverlay = []string{"wwinit", "wwclient"}
	b := a.Clone()
	b.SystemOverlay = []string{"wwinit", "wwclient", "foo"}

	changes := Diff(&a.Profile, &b.Profile)
	assert.Equal(t, []Change{{
		Path:   "system-overlays",
		Before: "[wwinit, wwclient]",
		After:  "[wwinit, wwclient, foo]",
	}}, changes)
}

func TestDiff_NestedKernelArgs(t *testing.T) {
	a := NewNode("n01")
	a.Kernel = &KernelConf{Args: []string{"quiet"}}
	b := a.Clone()
	b.Kernel.Args = []string{"quiet", "console=ttyS0"}

	changes := Diff(&a.Profile, &b.Profile)
	assert.Equal(t, []Change{{
		Path:   "kernel.kernelargs",
		Before: "[quiet]",
		After:  "[quiet, console=ttyS0]",
	}}, changes)
}

func TestDiff_PointerAutoCreated(t *testing.T) {
	a := NewNode("n01")
	b := a.Clone()
	b.Ipmi = &IpmiConf{UserName: "root"}

	changes := Diff(&a.Profile, &b.Profile)
	assert.Contains(t, changes, Change{Path: "ipmi.ipmiuser", Before: "<unset>", After: `"root"`})
}

func TestDiff_TagMap(t *testing.T) {
	a := NewNode("n01")
	a.Tags = map[string]string{"role": "compute"}
	b := a.Clone()
	b.Tags = map[string]string{"role": "gpu", "rack": "A1"}

	changes := Diff(&a.Profile, &b.Profile)
	assert.Contains(t, changes, Change{Path: "tags[role]", Before: `"compute"`, After: `"gpu"`})
	assert.Contains(t, changes, Change{Path: "tags[rack]", Before: "<unset>", After: `"A1"`})
}

func TestDiff_NetDevMap(t *testing.T) {
	a := NewNode("n01")
	a.NetDevs = map[string]*NetDev{
		"default": {Ipaddr: net.ParseIP("10.0.0.1"), Device: "eth0"},
	}
	b := a.Clone()
	b.NetDevs["default"].Ipaddr = net.ParseIP("10.0.0.2")
	b.NetDevs["secondary"] = &NetDev{Device: "eth1"}

	changes := Diff(&a.Profile, &b.Profile)
	assert.Contains(t, changes, Change{Path: "netdevs[default].ipaddr", Before: "10.0.0.1", After: "10.0.0.2"})
	assert.Contains(t, changes, Change{Path: "netdevs[secondary].netdev", Before: "<unset>", After: `"eth1"`})
}

func TestFormatChanges_Collapse(t *testing.T) {
	common := []Change{{Path: "system-overlays", Before: "[wwinit]", After: "[wwinit, foo]"}}
	uniq := []Change{{Path: "comment", Before: "<unset>", After: `"x"`}}

	out := FormatChanges(map[string][]Change{
		"n01": common,
		"n02": common,
		"n03": uniq,
	})

	assert.Equal(t, "n[01-02]:\n  system-overlays: [wwinit] → [wwinit, foo]\n\nn03:\n  comment: <unset> → \"x\"\n", out)
}

func TestFormatChanges_Empty(t *testing.T) {
	assert.Equal(t, "", FormatChanges(map[string][]Change{}))
	assert.Equal(t, "", FormatChanges(map[string][]Change{"n01": nil}))
}

// Diff over &Profile drops node-only fields (Discoverable, AssetKey). Diff
// over &Node must surface them so `wwctl node set --discoverable` and
// `--asset` show up in the confirmation summary.
func TestDiff_NodeOnlyFields(t *testing.T) {
	a := NewNode("n01")
	b := a.Clone()
	b.AssetKey = "rack-A1-slot-3"
	err := b.Discoverable.Set("true")
	assert.NoError(t, err)
	profileOnly := Diff(&a.Profile, &b.Profile)
	assert.Empty(t, profileOnly, "diffing &Profile must not see Node-level fields")

	changes := Diff(&a, b)
	assert.Contains(t, changes, Change{Path: "asset", Before: "<unset>", After: `"rack-A1-slot-3"`})
	var sawDiscoverable bool
	for _, c := range changes {
		if c.Path == "discoverable" {
			sawDiscoverable = true
			break
		}
	}
	assert.True(t, sawDiscoverable, "discoverable change must be reported")
}
