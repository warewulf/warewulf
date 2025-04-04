package api

import (
	"fmt"
	"net"
	"net/http"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func AuthMiddleware(auth *config.Authentication, allowedNets []net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wwlog.Debug("allowed subnets: %v", allowedNets)
			wwlog.Debug("remote address: %v", r.RemoteAddr)
			fromAllowedNet := false
			if ipStr, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				ip := net.ParseIP(ipStr)
				if ip == nil {
					http.Error(w, fmt.Sprintf("Invalid remote address: %v", r.RemoteAddr), http.StatusForbidden)
				}
				for _, allowedNet := range allowedNets {
					if allowedNet.Contains(ip) {
						fromAllowedNet = true
						break
					}
				}
				if !fromAllowedNet {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			} else {
				http.Error(w, fmt.Sprintf("Invalid remote address: %v", r.RemoteAddr), http.StatusForbidden)
				return
			}

			if auth != nil {
				username, password, ok := r.BasicAuth()
				if !ok {
					w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				_, err := auth.Authenticate(username, password)
				if err != nil {
					w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
