package hostlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Single(t *testing.T) {
	assert.Equal(t, []string{"node1"}, Expand([]string{"node1"}))
}

func Test_Multiple(t *testing.T) {
	assert.Equal(t, []string{"node1", "node2"}, Expand([]string{"node1", "node2"}))
}

func Test_Range(t *testing.T) {
	assert.Equal(t, []string{"node1", "node2"}, Expand([]string{"node[1-2]"}))
}

func Test_Internal_Comma(t *testing.T) {
	assert.Equal(t, []string{"node1", "node2"}, Expand([]string{"node[1,2]"}))
}

func Test_Mixed_Range_Comma(t *testing.T) {
	assert.Equal(t, []string{"node1", "node2", "node3"}, Expand([]string{"node[1,2-3]"}))
}

// not currently supported
//
// func Test_External_Comma(t *testing.T) {
// 	assert.Equal(t, []string{"node1", "node2"}, Expand([]string{"node1,node2"}))
// }
