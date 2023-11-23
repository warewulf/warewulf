package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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
			Id: Entry{
				value: []string{"n0000"},
			},
		},
	)
	assert.NoError(t, err)
}

func TestNodeDisk(t *testing.T) {
	node_config := `WW_INTERNAL: 43
nodes:
  n1:
    disks:
      /dev/vda:
        wipe_table: "true"
        partitions:
          scratch:
            should_exist: "true"
    filesystems:
      /dev/disk/by-partlabel/scratch:
        format: btrfs
        path: /scratch
        wipe_filesystem: "true"`

	config, parse_error := Parse([]byte(node_config))
	assert.Empty(t, parse_error)

	nodeInfos, info_error := config.FindAllNodes()
	assert.Empty(t, info_error)
	assert.Len(t, nodeInfos, 1)

	node := nodeInfos[0]
	assert.Len(t, node.Disks, 1)
	assert.Len(t, node.FileSystems, 1)

	disk := node.Disks["/dev/vda"]
	assert.True(t, disk.WipeTable.GetB())
	assert.Len(t, disk.Partitions, 1)

	partition := disk.Partitions["scratch"]
	assert.True(t, partition.ShouldExist.GetB())

	filesystem := node.FileSystems["/dev/disk/by-partlabel/scratch"]
	assert.Equal(t, "btrfs", filesystem.Format.Get())
	assert.Equal(t, "/scratch", filesystem.Path.Get())
	assert.True(t, filesystem.WipeFileSystem.GetB())
}
