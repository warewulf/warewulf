package warewulfd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_GetNodeOrSetDiscoverable(t *testing.T) {
	var tests = map[string]struct {
		nodesConf string
		hwaddr    string
		node      string
		err       bool
	}{
		"empty": {
			nodesConf: `
nodes: {}
`,
			hwaddr: "00:00:00:00:00:00",
			err:    true,
		},
		"configured": {
			nodesConf: `
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:01
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
		"discoverable": {
			nodesConf: `
nodes:
  n1:
    discoverable: true
    network devices:
      default: {}
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
		"discoverable with primary": {
			nodesConf: `
nodes:
  n1:
    discoverable: true
    primary netdev: default
    network devices:
      default: {}
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
		"discoverable without network": {
			nodesConf: `
nodeprofiles:
  default:
    network devices:
      default:
        netmask: 255.255.255.0
nodes:
  n1:
    profiles:
    - default
    discoverable: true
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("/etc/warewulf/nodes.conf", tt.nodesConf)

			err := LoadNodeDB()
			assert.NoError(t, err)

			node, err := GetNodeOrSetDiscoverable(tt.hwaddr)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.node, node.Id())
			}
		})
	}

}
