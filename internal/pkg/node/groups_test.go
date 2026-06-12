package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

func Test_GroupMembers_FromNodegroupsStanza(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01: {}
  n02: {}
  n03: {}
nodegroups:
  rack1:
    - n01
    - n02
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n02"}, registry.GroupMembers("rack1"))
}

func Test_GroupMembers_FromPerNodeField(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01:
    nodegroups:
      - admin
  n02: {}
  n03:
    nodegroups:
      - admin
      - rack2
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n03"}, registry.GroupMembers("admin"))
	assert.Equal(t, []string{"n03"}, registry.GroupMembers("rack2"))
}

func Test_GroupMembers_FromProfileInheritance(t *testing.T) {
	registry, err := Parse([]byte(`
nodeprofiles:
  gpu:
    nodegroups:
      - gpu-nodes
nodes:
  n01:
    profiles:
      - gpu
  n02: {}
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01"}, registry.GroupMembers("gpu-nodes"))
}

func Test_GroupMembers_UnionOfBothSources(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01:
    nodegroups:
      - admin
  n02: {}
  n03: {}
nodegroups:
  admin:
    - n02
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n02"}, registry.GroupMembers("admin"))
}

func Test_GroupMembers_HostlistRangeInNodegroupsStanza(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01: {}
  n02: {}
  n03: {}
  n04: {}
nodegroups:
  rack1:
    - n[01-03]
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n02", "n03"}, registry.GroupMembers("rack1"))
}

func Test_GroupMembers_AllBuiltin(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01: {}
  n02: {}
  n03: {}
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n02", "n03"}, registry.GroupMembers("all"))
}

func Test_GroupMembers_AllIgnoresUserDefinedAll(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01: {}
  n02: {}
nodegroups:
  all:
    - n01
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n02"}, registry.GroupMembers("all"))
}

func Test_GroupMembers_NegationViaProfileMerge(t *testing.T) {
	registry, err := Parse([]byte(`
nodeprofiles:
  base:
    nodegroups:
      - default
nodes:
  n01:
    profiles:
      - base
  n02:
    profiles:
      - base
    nodegroups:
      - ~default
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01"}, registry.GroupMembers("default"))
}

func Test_GroupMembers_All_HonorsLiteralOptOut(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01: {}
  n02:
    nodegroups:
      - ~all
  n03: {}
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n03"}, registry.GroupMembers("all"),
		"n02 carries literal `~all` and must be excluded from @all expansion")
}

func Test_GroupMembers_All_HonorsProfileInheritedOptOut(t *testing.T) {
	registry, err := Parse([]byte(`
nodeprofiles:
  quarantine:
    nodegroups:
      - ~all
nodes:
  n01: {}
  n02:
    profiles:
      - quarantine
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01"}, registry.GroupMembers("all"),
		"n02 inherits ~all from the quarantine profile and must be excluded")
}

func Test_GroupMembers_UnknownGroupIsEmpty(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01: {}
`))
	assert.NoError(t, err)
	assert.Empty(t, registry.GroupMembers("missing"))
}

func Test_HostlistExpand_WithGroupResolver(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01:
    nodegroups:
      - admin
  n02: {}
  n03: {}
  n04: {}
nodegroups:
  rack1:
    - n[01-02]
`))
	assert.NoError(t, err)
	hostlist.SetGroupResolver(&registry)
	t.Cleanup(func() { hostlist.SetGroupResolver(nil) })

	t.Run("plain hostlist", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n01", "n02"}, hostlist.Expand([]string{"n[01-02]"}))
	})

	t.Run("single group", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n01", "n02"}, hostlist.Expand([]string{"@rack1"}))
	})

	t.Run("mixed plain and group", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n01", "n03"}, hostlist.Expand([]string{"n03", "@admin"}))
	})

	t.Run("group dedupes against plain", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n01", "n02"}, hostlist.Expand([]string{"n01", "@rack1"}))
	})

	t.Run("@all returns every node", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n01", "n02", "n03", "n04"}, hostlist.Expand([]string{"@all"}))
	})

	t.Run("unknown group is ignored", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n02"}, hostlist.Expand([]string{"n02", "@bogus"}))
	})

	t.Run("empty @ token is ignored", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n02"}, hostlist.Expand([]string{"n02", "@"}))
	})

	t.Run("comma-separated group and plain", func(t *testing.T) {
		assert.ElementsMatch(t, []string{"n01", "n03"}, hostlist.Expand([]string{"n03,@admin"}))
	})
}

func Test_ListAllGroups(t *testing.T) {
	registry, err := Parse([]byte(`
nodeprofiles:
  gpu:
    nodegroups:
      - gpu-nodes
nodes:
  n01:
    nodegroups:
      - admin
  n02:
    profiles:
      - gpu
nodegroups:
  rack1:
    - n01
  rack2:
    - n02
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"admin", "gpu-nodes", "rack1", "rack2"}, registry.ListAllNodegroups())
}

func Test_ListNodesUsingGroup(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01:
    nodegroups:
      - admin
  n02: {}
nodegroups:
  admin:
    - n02
`))
	assert.NoError(t, err)
	nodes, err := registry.ListNodesUsingNodegroup("admin")
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	ids := []string{nodes[0].Id(), nodes[1].Id()}
	assert.ElementsMatch(t, []string{"n01", "n02"}, ids)
}
