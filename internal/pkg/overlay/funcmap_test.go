package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createIgnitionJson(t *testing.T) {
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

	expected_json := `{
  "ignition": {
    "version": "3.1.0"
  },
  "storage": {
    "disks": [
      {
        "device": "/dev/vda",
        "partitions": [
          {
            "label": "scratch",
            "shouldExist": true,
            "wipePartitionEntry": false
          }
        ],
        "wipeTable": true
      }
    ],
    "filesystems": [
      {
        "device": "/dev/disk/by-partlabel/scratch",
        "format": "btrfs",
        "path": "/scratch",
        "wipeFilesystem": true
      }
    ]
  }
}`

	config, parse_error := node.Parse([]byte(node_config))
	assert.Empty(t, parse_error)

	nodeInfos, info_error := config.FindAllNodes()
	assert.Empty(t, info_error)

	node := nodeInfos[0]
	assert.JSONEq(t, expected_json, createIgnitionJson(&node))
}
