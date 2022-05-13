package node

import (
	"testing"

	"gopkg.in/yaml.v2"
	"github.com/stretchr/testify/assert"
)

func TestNodeUpdate(t *testing.T) {
	var nodeConfig = `
WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
nodes:
  n0000:
    discoverable: true
    profiles:
    - default
    network devices:
      default:
        onboot: true
        device: eth0
        hwaddr: 08:00:27:39:46:70
        ipaddr: 10.0.8.150	
`
	var nodeYaml NodeYaml
	err := yaml.Unmarshal([]byte(nodeConfig), &nodeYaml)
	assert.NoError(t, err)

	err = nodeYaml.NodeUpdate(
		NodeInfo{
			Id: Entry {
				value: []string{"n0000"},
			},
		},
	)
	assert.NoError(t, err)
}