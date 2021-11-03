package warewulfd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_getHostPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/upper?word=abc", nil)
	w := httptest.NewRecorder()

	type args struct {
		w   http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name       string
		args       args
		RemoteAddr string
		hostIP     string
		port       int
		wantErr    bool
	}{
		{"IPv4", args{w, req}, "10.0.0.1:987", "10.0.0.1", 987, false},
		{"IPv4noPort", args{w, req}, "10.0.0.1", "", 0, true},
		{"IPv4noInt", args{w, req}, "10.0.0.1:foo", "10.0.0.1", 0, true},
		{"IPv6", args{w, req}, "[::ffff:192.0.2.128]:8080", "::ffff:192.0.2.128", 8080, false},
		{"IPv6noBrackets", args{w, req}, "::ffff:192.0.2.128:8080", "", 0, true},
		{"IPv6", args{w, req}, "[::ffff:192.0.2.128]:foo", "::ffff:192.0.2.128", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.req.RemoteAddr = tt.RemoteAddr
			got, got1, err := getHostPort(tt.args.w, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.hostIP {
				t.Errorf("getHostPort() got = %v, hostIP %v", got, tt.hostIP)
			}
			if got1 != tt.port {
				t.Errorf("getHostPort() got1 = %v, hostIP %v", got1, tt.port)
			}
		})
	}
}
