package overlay

import (
	"testing"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
)

var overlayImageTests = []struct{
	description string
	node string
	context string
	overlays []string
	image string
}{
	{"all empty", "", "", nil, ""},
	{"empty with named context", "", "system", nil, "p/overlays/__SYSTEM__.img"},
	{"empty with named overlay", "", "", []string{"o1"}, "p/overlays/o1.img"},
	{"empty with two named overlays", "", "", []string{"o1", "o2"}, "p/overlays/o1-o2.img"},
	{"empty node", "node1", "", nil, ""},
	{"node system overlay", "node1", "system", nil, "p/overlays/node1/__SYSTEM__.img"},
	{"node runtime overlay", "node1", "runtime", nil, "p/overlays/node1/__RUNTIME__.img"},
	{"node single overlay", "node1", "", []string{"o1"}, "p/overlays/node1/o1.img"},
	{"node two overlays", "node1", "", []string{"o1", "o2"}, "p/overlays/node1/o1-o2.img"},
	{"node with context and overlays", "node1", "system", []string{"o1", "o2"}, "p/overlays/node1/__SYSTEM__.img"},
}

func Test_OverlayImage(t *testing.T) {
	conf := warewulfconf.Get()
	conf.Paths.WWProvisiondir = "p"
	for _, tt := range overlayImageTests {
		t.Run(tt.description, func(t *testing.T) {
			out := OverlayImage(tt.node, tt.context, tt.overlays)
			if  tt.image != out {
				t.Errorf("got %q, want %q", out, tt.image)
			}
		})
	}
}
