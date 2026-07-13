package nodestatus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

func TestStatusURL(t *testing.T) {
	tests := map[string]struct {
		ipaddr  string
		ipaddr6 string
		port    int
		want    string
		wantErr bool
	}{
		"ipv4": {
			ipaddr: "10.0.0.1",
			port:   9873,
			want:   "http://10.0.0.1:9873/status",
		},
		"ipv6 only": {
			ipaddr6: "2001:db8::1",
			port:    9873,
			want:    "http://[2001:db8::1]:9873/status",
		},
		"ipv4 preferred over ipv6": {
			ipaddr:  "10.0.0.1",
			ipaddr6: "2001:db8::1",
			port:    9873,
			want:    "http://10.0.0.1:9873/status",
		},
		"ipv6 loopback": {
			ipaddr6: "::1",
			port:    9873,
			want:    "http://[::1]:9873/status",
		},
		"neither configured": {
			port:    9873,
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			controller := warewulfconf.New()
			controller.Ipaddr = tt.ipaddr
			controller.Ipaddr6 = tt.ipaddr6
			controller.Warewulf.Port = tt.port

			got, err := statusURL(controller)
			if tt.wantErr {
				assert.ErrorContains(t, err, "ipaddr6")
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
