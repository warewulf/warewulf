package configure

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// Test_buildHostOverlays covers the cases in which a service builds
// nothing. Each must return before buildHostOverlays reaches an overlay
// build, since those write into the Warewulf server's own /.
func Test_buildHostOverlays(t *testing.T) {
	tests := map[string]struct {
		enableHostOverlay bool
		overlays          warewulfconf.OverlayList
		log               string
	}{
		// An in-place upgrade leaves warewulf.conf without any overlays
		// keys, because it is packaged %config(noreplace). Say so by name:
		// the service is restarted either way, so silence here looks like
		// success.
		"no overlays configured": {
			enableHostOverlay: true,
			overlays:          nil,
			log:               "No host overlays configured for dhcp",
		},
		"host overlays disabled": {
			enableHostOverlay: false,
			overlays:          warewulfconf.OverlayList{"dhcpd"},
			log:               "host overlays are disabled, did not configure dhcp",
		},
		"host overlays disabled and none configured": {
			enableHostOverlay: false,
			overlays:          nil,
			log:               "host overlays are disabled, did not configure dhcp",
		},
		"named overlay does not exist": {
			enableHostOverlay: true,
			overlays:          warewulfconf.OverlayList{"nosuchoverlay"},
			log:               "Skipping host overlay nosuchoverlay",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			conf := env.Configure()
			conf.Warewulf.EnableHostOverlayP = &tt.enableHostOverlay

			logbuf := bytes.NewBufferString("")
			wwlog.SetLogWriter(logbuf)
			wwlog.SetLogLevel(wwlog.INFO)

			assert.NoError(t, buildHostOverlays("dhcp", tt.overlays))
			assert.Contains(t, logbuf.String(), tt.log)
		})
	}
}
