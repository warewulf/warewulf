package node

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	nodeConfYml1 := `
WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
  test:
    comment: This is a test profile
nodes:
  n01:
    discoverable: true
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.0.10.1
  n02:
    discoverable: true
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.0.10.2
`
	nodeConfYml2 := `
WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
  test:
    comment: This is a test profile
nodes:
  n02:
    discoverable: true
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.0.10.2
  n01:
    discoverable: true
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.0.10.1
`
	nodeConfYml3 := `
WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
  test:
    comment: This is a test profile
nodes:
  n02:
    discoverable: true
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.0.10.2
  n01:
    discoverable: true
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.0.10.3
`
	var nodeConf1, nodeConf2, nodeConf3 NodeYaml
	err := yaml.Unmarshal([]byte(nodeConfYml1), &nodeConf1)
	assert.NoError(t, err)
	err = yaml.Unmarshal([]byte(nodeConfYml2), &nodeConf2)
	assert.NoError(t, err)
	err = yaml.Unmarshal([]byte(nodeConfYml3), &nodeConf3)
	assert.NoError(t, err)

	t.Run("Same NodeYaml with same conf", func(t *testing.T) {
		var testConf NodeYaml
		err = yaml.Unmarshal([]byte(nodeConfYml1), &testConf)
		assert.NoError(t, err)
		if testConf.Hash() != nodeConf1.Hash() {
			err = yaml.Unmarshal([]byte(nodeConfYml1), nodeConf1)
			assert.NoError(t, err)
			t.Errorf("Hashes for same configuration differs: %x != %x", nodeConf1.Hash(), nodeConf1.Hash())
		}
	})

	t.Run("Different sorted NodeYaml with same conf", func(t *testing.T) {
		yml1, err := yaml.Marshal(nodeConf1)
		assert.NoError(t, err)
		yml2, err := yaml.Marshal(nodeConf2)
		assert.NoError(t, err)
		if nodeConf2.Hash() != nodeConf1.Hash() {
			t.Errorf("Hashes for same configuration differs: %x != %x\njson1:\n%s,yml2:\n%s\n", nodeConf2.Hash(), nodeConf1.Hash(), yml1, yml2)
		}
	})

	t.Run("Different NodeYaml with different conf", func(t *testing.T) {
		if nodeConf2.Hash() == nodeConf3.Hash() {
			t.Errorf("Hashes for different configuration is the same: %x == %x", nodeConf2.Hash(), nodeConf3.Hash())
		}
	})
}
