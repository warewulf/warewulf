package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_MergeNode(t *testing.T) {
	var tests = map[string]struct {
		nodesConf string
		node      string
		field     string
		source    string
		value     string
		nodes     []string
		fields    []string
		sources   []string
		values    []string
	}{
		"node comment": {
			nodesConf: `
nodes:
  n1:
    comment: n1 comment`,
			node:   "n1",
			field:  "Comment",
			source: "",
			value:  "n1 comment",
		},
		"profile comment": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
nodeprofiles:
  p1:
    comment: p1 comment`,
			node:   "n1",
			field:  "Comment",
			source: "p1",
			value:  "p1 comment",
		},
		"multiple profile comments": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    comment: p1 comment
  p2:
    comment: p2 comment`,
			node:   "n1",
			field:  "Comment",
			source: "p2",
			value:  "p2 comment",
		},
		"node comment supersedes profile comment": {
			nodesConf: `
nodes:
  n1:
    comment: n1 comment
    profiles:
    - p1
nodeprofiles:
  p1:
    comment: p1 comment`,
			node:   "n1",
			field:  "Comment",
			source: "SUPERSEDED",
			value:  "n1 comment",
		},
		"node comment supersedes multiple profile comments": {
			nodesConf: `
nodes:
  n1:
    comment: n1 comment
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    comment: p1 comment
  p2:
    comment: p2 comment`,
			node:   "n1",
			field:  "Comment",
			source: "SUPERSEDED",
			value:  "n1 comment",
		},
		"node kernel args": {
			nodesConf: `
nodes:
  n1:
    kernel:
      args: n1 args`,
			node:   "n1",
			field:  "Kernel.Args",
			source: "",
			value:  "n1 args",
		},
		"profile kernel args": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
nodeprofiles:
  p1:
    kernel:
      args: p1 args`,
			node:   "n1",
			field:  "Kernel.Args",
			source: "p1",
			value:  "p1 args",
		},
		"multiple profile kernel args": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    kernel:
      args: p1 args
  p2:
    kernel:
      args: p2 args`,
			node:   "n1",
			field:  "Kernel.Args",
			source: "p2",
			value:  "p2 args",
		},
		"node kernel args supersedes profile kernel args": {
			nodesConf: `
nodes:
  n1:
    kernel:
      args: n1 args
    profiles:
    - p1
nodeprofiles:
  p1:
    kernel:
      args: p1 args`,
			node:   "n1",
			field:  "Kernel.Args",
			source: "SUPERSEDED",
			value:  "n1 args",
		},
		"node kernel args supersedes multiple profile kernel args": {
			nodesConf: `
nodes:
  n1:
    kernel:
      args: n1 args
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    kernel:
      args: p1 args
  p2:
    kernel:
      args: p2 args`,
			node:   "n1",
			field:  "Kernel.Args",
			source: "SUPERSEDED",
			value:  "n1 args",
		},
		"node tag": {
			nodesConf: `
nodes:
  n1:
    tags:
      tag: n1 tag`,
			node:   "n1",
			field:  "Tags[tag]",
			source: "",
			value:  "n1 tag",
		},
		"profile tag": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
nodeprofiles:
  p1:
    tags:
      tag: p1 tag`,
			node:   "n1",
			field:  "Tags[tag]",
			source: "p1",
			value:  "p1 tag",
		},
		"multiple profile tags": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    tags:
      tag: p1 tag
  p2:
    tags:
      tag: p2 tag`,
			node:   "n1",
			field:  "Tags[tag]",
			source: "p2",
			value:  "p2 tag",
		},
		"node tag supersedes profile tag": {
			nodesConf: `
nodes:
  n1:
    tags:
      tag: n1 tag
    profiles:
    - p1
nodeprofiles:
  p1:
    tags:
      tag: p1 tag`,
			node:   "n1",
			field:  "Tags[tag]",
			source: "SUPERSEDED",
			value:  "n1 tag",
		},
		"node tag supersedes multiple profile tags": {
			nodesConf: `
nodes:
  n1:
    tags:
      tag: n1 tag
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    tags:
      tag: p1 tag
  p2:
    tags:
      tag: p2 tag`,
			node:   "n1",
			field:  "Tags[tag]",
			source: "SUPERSEDED",
			value:  "n1 tag",
		},
		"mixture of tags from nodes and profiles": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
    tags:
      n1: n1 tag
nodeprofiles:
  p1:
    tags:
      p1: p1 tag
  p2:
    tags:
      p2: p2 tag`,
			nodes: []string{
				"n1",
				"n1",
				"n1",
			},
			fields: []string{
				"Tags[n1]",
				"Tags[p1]",
				"Tags[p2]",
			},
			sources: []string{
				"",
				"p1",
				"p2",
			},
			values: []string{
				"n1 tag",
				"p1 tag",
				"p2 tag",
			},
		},
		"node system overlay": {
			nodesConf: `
nodes:
  n1:
    system overlay:
    - no1
    - no2`,
			node:   "n1",
			field:  "SystemOverlay",
			source: "",
			value:  "no1,no2",
		},
		"profile system overlay": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
nodeprofiles:
  p1:
    system overlay:
    - po1
    - po2`,
			node:   "n1",
			field:  "SystemOverlay",
			source: "p1",
			value:  "po1,po2",
		},
		"multiple profile system overlays": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    system overlay:
    - po1
    - po2
  p2:
    system overlay:
    - po3
    - po4`,
			node:   "n1",
			field:  "SystemOverlay",
			source: "p1,p2",
			value:  "po1,po2,po3,po4",
		},
		"node system overlay adds to profile system overlay": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    system overlay:
    - no1
    - no2
nodeprofiles:
  p1:
    system overlay:
    - po1
    - po2`,
			node:   "n1",
			field:  "SystemOverlay",
			source: "p1,n1",
			value:  "po1,po2,no1,no2",
		},
		"node system overlay adds to multiple profile system overlays": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
    system overlay:
    - no1
    - no2
nodeprofiles:
  p1:
    system overlay:
    - po1
    - po2
  p2:
    system overlay:
    - po3
    - po4`,
			node:   "n1",
			field:  "SystemOverlay",
			source: "p1,p2,n1",
			value:  "po1,po2,po3,po4,no1,no2",
		},
		"node netdev tag": {
			nodesConf: `
nodes:
  n1:
    network devices:
      default:
        tags:
          tag: n1 netdev tag`,
			node:   "n1",
			field:  "NetDevs[default].Tags[tag]",
			source: "",
			value:  "n1 netdev tag",
		},
		"profile netdev tag": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
nodeprofiles:
  p1:
    network devices:
      default:
        tags:
          tag: p1 netdev tag`,
			node:   "n1",
			field:  "NetDevs[default].Tags[tag]",
			source: "p1",
			value:  "p1 netdev tag",
		},
		"multiple profile netdev tags": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
nodeprofiles:
  p1:
    network devices:
      default:
        tags:
          tag: p1 netdev tag
  p2:
    network devices:
      default:
        tags:
          tag: p2 netdev tag`,
			node:   "n1",
			field:  "NetDevs[default].Tags[tag]",
			source: "p2",
			value:  "p2 netdev tag",
		},
		"node supercededs profile netdev tag": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    network devices:
      default:
        tags:
          tag: n1 netdev tag
nodeprofiles:
  p1:
    network devices:
      default:
        tags:
          tag: p1 netdev tag`,
			node:   "n1",
			field:  "NetDevs[default].Tags[tag]",
			source: "SUPERSEDED",
			value:  "n1 netdev tag",
		},
		"node supersedes multiple profile netdev tags": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
    network devices:
      default:
        tags:
          tag: n1 netdev tag
nodeprofiles:
  p1:
    network devices:
      default:
        tags:
          tag: p1 netdev tag
  p2:
    network devices:
      default:
        tags:
          tag: p2 netdev tag`,
			node:   "n1",
			field:  "NetDevs[default].Tags[tag]",
			source: "SUPERSEDED",
			value:  "n1 netdev tag",
		},
		"mixture of netdev tags from nodes and profiles": {
			nodesConf: `
nodes:
  n1:
    profiles:
    - p1
    - p2
    network devices:
      default:
        tags:
          n1: n1 netdev tag
nodeprofiles:
  p1:
    network devices:
      default:
        tags:
          p1: p1 netdev tag
  p2:
    network devices:
      default:
        tags:
          p2: p2 netdev tag`,
			nodes: []string{
				"n1",
				"n1",
				"n1",
			},
			fields: []string{
				"NetDevs[default].Tags[n1]",
				"NetDevs[default].Tags[p1]",
				"NetDevs[default].Tags[p2]",
			},
			sources: []string{
				"",
				"p1",
				"p2",
			},
			values: []string{
				"n1 netdev tag",
				"p1 netdev tag",
				"p2 netdev tag",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll(t)
			env.WriteFile(t, "/etc/warewulf/nodes.conf", tt.nodesConf)

			registry, regErr := New()
			assert.NoError(t, regErr)

			if tt.node != "" {
				node, fields, mergeErr := registry.MergeNode(tt.node)
				assert.NoError(t, mergeErr)

				value, valueErr := getNestedFieldString(node, tt.field)
				assert.NoError(t, valueErr)
				assert.Equal(t, tt.value, value)
				assert.Equal(t, tt.value, fields.Value(tt.field))
				assert.Equal(t, tt.source, fields.Source(tt.field))
			}

			var nodes []Node
			for i, _ := range tt.nodes {
				node, _, mergeErr := registry.MergeNode(tt.nodes[i])
				assert.NoError(t, mergeErr)
				nodes = append(nodes, node)
			}

			for i, _ := range tt.nodes {
				_, fields, _ := registry.MergeNode(tt.nodes[i])
				value, valueErr := getNestedFieldString(nodes[i], tt.fields[i])
				assert.NoError(t, valueErr)
				assert.Equal(t, tt.values[i], value)
				assert.Equal(t, tt.values[i], fields.Value(tt.fields[i]))
				assert.Equal(t, tt.sources[i], fields.Source(tt.fields[i]))
			}
		})
	}
}
