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

func TestCompress(t *testing.T) {
	tests := map[string]struct {
		input  []string
		output string
	}{
		"empty":             {input: nil, output: ""},
		"single":            {input: []string{"n01"}, output: "n01"},
		"consecutive":       {input: []string{"n01", "n02", "n03"}, output: "n[01-03]"},
		"sparse":            {input: []string{"n01", "n03", "n05"}, output: "n[01,03,05]"},
		"mixed run + gap":   {input: []string{"n01", "n02", "n04"}, output: "n[01-02,04]"},
		"different prefix":  {input: []string{"n01", "n02", "rack01"}, output: "n[01-02],rack01"},
		"different width":   {input: []string{"n1", "n01"}, output: "n1,n01"},
		"no digits":         {input: []string{"head", "login"}, output: "head,login"},
		"cross-zero rollup": {input: []string{"n09", "n10", "n11"}, output: "n[09-11]"},
		"out of order":      {input: []string{"n03", "n01", "n02"}, output: "n[01-03]"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.output, Compress(tt.input))
		})
	}
}
