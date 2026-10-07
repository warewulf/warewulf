package overlay

import (
	"bytes"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// Test_hostOverlaysToBuild covers which host overlays a service ends up
// applying. Notably, a legacy `host` overlay left on disk is applied only
// when a service names it: it used to be appended to every service's list,
// which meant each wwctl configure subcommand rewrote every file the
// legacy overlay held, and did so at the highest precedence.
//
// This tests the selection step rather than BuildHostOverlay itself,
// because that builds into / and so cannot be exercised safely here.
func Test_hostOverlaysToBuild(t *testing.T) {
	tests := map[string]struct {
		onDisk      []string
		overlayName []string
		build       []string
	}{
		"legacy host overlay is not implied": {
			onDisk:      []string{"dhcpd", "host"},
			overlayName: []string{"dhcpd"},
			build:       []string{"dhcpd"},
		},
		"legacy host overlay is not implied for an empty list": {
			onDisk:      []string{"host"},
			overlayName: nil,
			build:       nil,
		},
		"legacy host overlay is applied when named": {
			onDisk:      []string{"dhcpd", "host"},
			overlayName: []string{"dhcpd", "host"},
			build:       []string{"dhcpd", "host"},
		},
		"order is preserved": {
			onDisk:      []string{"dnsmasq", "tftproot"},
			overlayName: []string{"dnsmasq", "tftproot"},
			build:       []string{"dnsmasq", "tftproot"},
		},
		"a missing overlay is skipped, not an error": {
			onDisk:      []string{"dhcpd"},
			overlayName: []string{"dhcpd", "nosuchoverlay"},
			build:       []string{"dhcpd"},
		},
		"a missing legacy host overlay is skipped": {
			onDisk:      []string{"dhcpd"},
			overlayName: []string{"dhcpd", "host"},
			build:       []string{"dhcpd"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			for _, overlayName := range tt.onDisk {
				env.WriteFile(
					path.Join("var/lib/warewulf/overlays", overlayName, "rootfs/etc/testfile.ww"),
					"# test overlay")
			}

			build, err := hostOverlaysToBuild(tt.overlayName)
			assert.NoError(t, err)
			assert.Equal(t, tt.build, build)
		})
	}
}

// Test_checkHostOverlay_Permissions covers when applying an overlay to the
// Warewulf server warrants a security warning. A host overlay is written
// into / as root, so the hazard is that someone other than root can write
// to it and thereby choose what gets written. That is group- or
// other-writability, and nothing else: a mode may differ from the 0750 the
// packaged host overlays ship with and still be perfectly safe.
func Test_checkHostOverlay_Permissions(t *testing.T) {
	tests := map[string]struct {
		mode int
		warn bool
	}{
		"0750, as the packaged host overlays ship": {mode: 0o750, warn: false},
		"0700": {mode: 0o700, warn: false},
		"0755, as an overlay shared with nodes ships": {mode: 0o755, warn: false},
		"0555, stricter on writes than 0750":          {mode: 0o555, warn: false},
		"0775, root-owned but group-writable":         {mode: 0o775, warn: true},
		"0757, other-writable":                        {mode: 0o757, warn: true},
		"0777":                                        {mode: 0o777, warn: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("var/lib/warewulf/overlays/o1/rootfs/etc/testfile.ww", "# test overlay")
			env.Chmod("var/lib/warewulf/overlays/o1/rootfs", tt.mode)
			// Deferred LIFO, so this runs before RemoveAll: a mode
			// without owner write (0555) cannot be cleaned up as is.
			defer env.Chmod("var/lib/warewulf/overlays/o1/rootfs", 0o755)

			logbuf := bytes.NewBufferString("")
			wwlog.SetLogWriter(logbuf)

			ok, err := checkHostOverlay("o1")
			assert.NoError(t, err)
			assert.True(t, ok)

			if tt.warn {
				assert.Contains(t, logbuf.String(), "writable by group or other")
				assert.Contains(t, logbuf.String(), "o1")
			} else {
				assert.NotContains(t, logbuf.String(), "writable by group or other")
			}
		})
	}
}
