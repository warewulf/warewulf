package warewulfd

import (
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/stretchr/testify/assert"
	"testing"
)

var getOverlayFileTests = []struct {
	description string
	node        string
	context     string
	overlays    []string
	result      string
	succeed     bool // getOverlayFile will always fail if no overlay dir is defined!
}{
	{"empty", "", "", nil, "", false},
	{"empty node", "node1", "", nil, "", false},
	{"specific overlays without node", "", "", []string{"o1", "o2"}, "p/overlays/node1/o1-o2.img", false},
	{"system overlay", "node1", "system", nil, "p/overlays/node1/__SYSTEM__.img", false},
	{"runtime overlay", "node1", "runtime", nil, "p/overlays/node1/__RUNTIME__.img", false},
	{"specific overlay", "node1", "", []string{"o1"}, "p/overlays/node1/o1.img", false},
	{"multiple specific overlays", "node1", "", []string{"o1", "o2"}, "p/overlays/node1/o1-o2.img", false},
	{"multiple specific overlays with context", "node1", "system", []string{"o1", "o2"}, "p/overlays/node1/__SYSTEM__.img", false},
}

func Test_getOverlayFile(t *testing.T) {
	wwlog.SetLogLevel(wwlog.VERBOSE)
	conf := warewulfconf.Get()
	conf.Paths.WWProvisiondir = "p"
	var n node.NodeInfo
	for _, tt := range getOverlayFileTests {
		n.Id.Set(tt.node)
		t.Run(tt.description, func(t *testing.T) {
			result, err := getOverlayFile(n, tt.context, tt.overlays, false)
			if !tt.succeed {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.result, result)
		})
	}
}
