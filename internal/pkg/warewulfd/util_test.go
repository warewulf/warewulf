package warewulfd

import (
	"testing"
	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
)

var getOverlayFileTests = []struct{
	description string
	node string
	context string
	overlays []string
	result string
	succeed bool
}{
	{"empty", "", "", nil, "", true},
	{"empty node", "node1", "", nil, "", true},
	{"specific overlays without node", "", "", []string{"o1", "o2"}, "p/overlays/o1-o2.img", true},
	{"system overlay", "node1", "system", nil, "p/overlays/node1/__SYSTEM__.img", true},
	{"runtime overlay", "node1", "runtime", nil, "p/overlays/node1/__RUNTIME__.img", true},
	{"specific overlay", "node1", "", []string{"o1"}, "p/overlays/node1/o1.img", true},
	{"multiple specific overlays", "node1", "", []string{"o1", "o2"}, "p/overlays/node1/o1-o2.img", true},
	{"multiple specific overlays with context", "node1", "system", []string{"o1", "o2"}, "p/overlays/node1/__SYSTEM__.img", true},
}

func Test_getOverlayFile(t *testing.T) {
	conf := warewulfconf.Get()
	conf.Paths.WWProvisiondir = "p"
	for _, tt := range getOverlayFileTests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := getOverlayFile(tt.node, tt.context, tt.overlays, false)
			if !tt.succeed {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.result, result)
		})
	}
}
