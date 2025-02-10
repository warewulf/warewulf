package hostlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostList(t *testing.T) {
	tests := map[string]struct {
		input  []string
		output []string
	}{
		"single": {
			input:  []string{"node1"},
			output: []string{"node1"},
		},
		"multiple": {
			input:  []string{"node1", "node2"},
			output: []string{"node1", "node2"},
		},
		"range": {
			input:  []string{"node[1-2]"},
			output: []string{"node1", "node2"},
		},
		"internal comma": {
			input:  []string{"node[1,2]"},
			output: []string{"node1", "node2"},
		},
		"mixed range comma": {
			input:  []string{"node[1,2-3]"},
			output: []string{"node1", "node2", "node3"},
		},
		"external comma": {
			input:  []string{"node1,node2"},
			output: []string{"node1", "node2"},
		},
		"mixed external comma with range": {
			input:  []string{"n[1-3],n5,n[7-8,10]"},
			output: []string{"n1", "n2", "n3", "n5", "n7", "n8", "n10"},
		},
		"leading zeroes": {
			input:  []string{"n[01-03]"},
			output: []string{"n01", "n02", "n03"},
		},
		"double expansion": {
			input:  []string{"r[1-2]-n[3-4]"},
			output: []string{"r1-n3", "r1-n4", "r2-n3", "r2-n4"},
		},
		"double expansion, with commas": {
			input:  []string{"r[1,2]-n[3,4]"},
			output: []string{"r1-n3", "r1-n4", "r2-n3", "r2-n4"},
		},
		"wrong comma order": {
			input:  []string{"node[4,1]"},
			output: []string{"node4", "node1"},
		},
		"wrong dash order": {
			input:  []string{"node[4-1]"},
			output: nil,
		},
		"minus node": {
			input:  []string{"node[-1]"},
			output: []string{"node0", "node1"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.output, Expand(tt.input))
		})
	}
}
