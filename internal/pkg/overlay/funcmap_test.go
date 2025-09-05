package overlay

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func Test_createIgnitionJson(t *testing.T) {
	node_config := `nodes:
  n1:
    disks:
      /dev/vda:
        wipe_table: true
        partitions:
          scratch:
            should_exist: true
    filesystems:
      /dev/disk/by-partlabel/scratch:
        format: btrfs
        path: /scratch
        wipe_filesystem: true`

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
            "shouldExist": true
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

func Test_UniqueField(t *testing.T) {
	tests := map[string]struct {
		input  string
		sep    string
		field  int
		output string
	}{
		"empty": {
			input:  ``,
			sep:    ":",
			field:  0,
			output: ``,
		},

		"unique input": {
			input: `
name1:aaaa
name2:bbbb
`,
			sep:   ":",
			field: 0,
			output: `
name1:aaaa
name2:bbbb
`,
		},

		"duplicate field": {
			input: `
name1: aaaa
name1: bbbb
`,
			sep:   ":",
			field: 0,
			output: `
name1: aaaa
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.output, UniqueField(tt.sep, tt.field, tt.input))
		})
	}
}
