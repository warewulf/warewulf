package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ListAllNodes(t *testing.T) {
	tests := map[string]struct {
		registry string
		nodes    []string
	}{
		"empty": {
			registry: ``,
			nodes:    []string{},
		},
		"one node": {
			registry: `
nodes:
  n1: {}`,
			nodes: []string{"n1"},
		},
		"multiple nodes": {
			registry: `
nodes:
  n1: {}
  n2: {}`,
			nodes: []string{"n1", "n2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.nodes, registry.ListAllNodes())
		})
	}
}

func Test_ListAllProfiles(t *testing.T) {
	tests := map[string]struct {
		registry string
		profiles []string
	}{
		"empty": {
			registry: ``,
			profiles: []string{},
		},
		"one profile": {
			registry: `
nodeprofiles:
  p1: {}`,
			profiles: []string{"p1"},
		},
		"multiple profiles": {
			registry: `
nodeprofiles:
  p1: {}
  p2: {}`,
			profiles: []string{"p1", "p2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.profiles, registry.ListAllProfiles())
		})
	}
}

func Test_ListNodesUsingProfile(t *testing.T) {
	tests := map[string]struct {
		registry string
		profile  string
		nodes    []string
	}{
		"empty": {
			registry: ``,
			profile:  "p1",
			nodes:    []string{},
		},
		"node without profle": {
			registry: `
nodes:
  n1: {}`,
			profile: "p1",
			nodes:   []string{},
		},
		"node with profle": {
			registry: `
nodes:
  n1:
    profiles:
      - p1`,
			profile: "p1",
			nodes:   []string{"n1"},
		},
		"multiple nodes one with profile": {
			registry: `
nodes:
  n1: {}
  n2:
    profiles:
      - p1`,
			profile: "p1",
			nodes:   []string{"n2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.nodes, registry.ListNodesUsingProfile(tt.profile))
		})
	}
}

func Test_ListProfilesUsingProfile(t *testing.T) {
	tests := map[string]struct {
		registry string
		profile  string
		profiles []string
	}{
		"empty": {
			registry: ``,
			profile:  "p1",
			profiles: []string{},
		},
		"profile without profle": {
			registry: `
nodeprofiles:
  p1: {}`,
			profile:  "p2",
			profiles: []string{},
		},
		"profile with profile": {
			registry: `
nodeprofiles:
  p1:
    profiles:
      - p2`,
			profile:  "p2",
			profiles: []string{"p1"},
		},
		"multiple profiles one with profile": {
			registry: `
nodeprofiles:
  p1: {}
  p2:
    profiles:
      - p3`,
			profile:  "p3",
			profiles: []string{"p2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.profiles, registry.ListProfilesUsingProfile(tt.profile))
		})
	}
}

func Test_ListNodesUsingImage(t *testing.T) {
	tests := map[string]struct {
		registry string
		image    string
		nodes    []string
	}{
		"empty": {
			registry: ``,
			image:    "i1",
			nodes:    []string{},
		},
		"node without image": {
			registry: `
nodes:
  n1: {}`,
			image: "i1",
			nodes: []string{},
		},
		"node with image": {
			registry: `
nodes:
  n1:
    image name: i1`,
			image: "i1",
			nodes: []string{"n1"},
		},
		"multiple nodes one with profile": {
			registry: `
nodes:
  n1: {}
  n2:
    image name: i1`,
			image: "i1",
			nodes: []string{"n2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.nodes, registry.ListNodesUsingImage(tt.image))
		})
	}
}

func Test_ListProfilesUsingImage(t *testing.T) {
	tests := map[string]struct {
		registry string
		image    string
		profiles []string
	}{
		"empty": {
			registry: ``,
			image:    "i1",
			profiles: []string{},
		},
		"profile without image": {
			registry: `
nodeprofiles:
  p1: {}`,
			image:    "i1",
			profiles: []string{},
		},
		"profile with image": {
			registry: `
nodeprofiles:
  p1:
    image name: i1`,
			image:    "i1",
			profiles: []string{"p1"},
		},
		"multiple profiles one with image": {
			registry: `
nodeprofiles:
  p1: {}
  p2:
    image name: i1`,
			image:    "i1",
			profiles: []string{"p2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.profiles, registry.ListProfilesUsingImage(tt.image))
		})
	}
}

func Test_ListNodesUsingOverlay(t *testing.T) {
	tests := map[string]struct {
		registry string
		overlay  string
		nodes    []string
	}{
		"empty": {
			registry: ``,
			overlay:  "o1",
			nodes:    []string{},
		},
		"node without profle": {
			registry: `
nodes:
  n1: {}`,
			overlay: "o1",
			nodes:   []string{},
		},
		"node with runtime overlay": {
			registry: `
nodes:
  n1:
    runtime overlay:
      - o1`,
			overlay: "o1",
			nodes:   []string{"n1"},
		},
		"multiple nodes one with runtime overlay": {
			registry: `
nodes:
  n1: {}
  n2:
    runtime overlay:
      - o1`,
			overlay: "o1",
			nodes:   []string{"n2"},
		},
		"node with system overlay": {
			registry: `
nodes:
  n1:
    system overlay:
      - o1`,
			overlay: "o1",
			nodes:   []string{"n1"},
		},
		"multiple nodes one with system overlay": {
			registry: `
nodes:
  n1: {}
  n2:
    system overlay:
      - o1`,
			overlay: "o1",
			nodes:   []string{"n2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.nodes, registry.ListNodesUsingOverlay(tt.overlay))
		})
	}
}

func Test_ListProfilesUsingOverlay(t *testing.T) {
	tests := map[string]struct {
		registry string
		overlay  string
		profiles []string
	}{
		"empty": {
			registry: ``,
			overlay:  "o1",
			profiles: []string{},
		},
		"node without profle": {
			registry: `
nodeprofiles:
  p1: {}`,
			overlay:  "o1",
			profiles: []string{},
		},
		"node with runtime overlay": {
			registry: `
nodeprofiles:
  p1:
    runtime overlay:
      - o1`,
			overlay:  "o1",
			profiles: []string{"p1"},
		},
		"multiple nodes one with runtime overlay": {
			registry: `
nodeprofiles:
  p1: {}
  p2:
    runtime overlay:
      - o1`,
			overlay:  "o1",
			profiles: []string{"p2"},
		},
		"node with system overlay": {
			registry: `
nodeprofiles:
  p1:
    system overlay:
      - o1`,
			overlay:  "o1",
			profiles: []string{"p1"},
		},
		"multiple nodes one with system overlay": {
			registry: `
nodeprofiles:
  p1: {}
  p2:
    system overlay:
      - o1`,
			overlay:  "o1",
			profiles: []string{"p2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry, err := Parse([]byte(tt.registry))
			assert.NoError(t, err)
			assert.Equal(t, tt.profiles, registry.ListProfilesUsingOverlay(tt.overlay))
		})
	}
}
