package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/hostlist"
)

func Test_GroupMembers_FromPerNodeField(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01:
    groups:
      - admin
  n02: {}
  n03:
    groups:
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
    groups:
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

func Test_GroupMembers_UnionOfNodeAndProfile(t *testing.T) {
	registry, err := Parse([]byte(`
nodeprofiles:
  base:
    groups:
      - default
nodes:
  n01:
    profiles:
      - base
  n02:
    groups:
      - default
  n03: {}
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"n01", "n02"}, registry.GroupMembers("default"))
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

func Test_GroupMembers_NegationViaProfileMerge(t *testing.T) {
	registry, err := Parse([]byte(`
nodeprofiles:
  base:
    groups:
      - default
nodes:
  n01:
    profiles:
      - base
  n02:
    profiles:
      - base
    groups:
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
    groups:
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
    groups:
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
    groups:
      - admin
      - rack1
  n02:
    groups:
      - rack1
  n03: {}
  n04: {}
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
    groups:
      - gpu-nodes
nodes:
  n01:
    groups:
      - admin
      - rack1
  n02:
    profiles:
      - gpu
    groups:
      - rack2
`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"admin", "gpu-nodes", "rack1", "rack2"}, registry.ListAllGroups())
}

func Test_ListNodesUsingGroup(t *testing.T) {
	registry, err := Parse([]byte(`
nodes:
  n01:
    groups:
      - admin
  n02:
    groups:
      - admin
`))
	assert.NoError(t, err)
	nodes, err := registry.ListNodesUsingGroup("admin")
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	ids := []string{nodes[0].Id(), nodes[1].Id()}
	assert.ElementsMatch(t, []string{"n01", "n02"}, ids)
}
