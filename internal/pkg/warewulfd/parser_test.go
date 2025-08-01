package warewulfd

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var parseReqTests = []struct {
	description string
	url         string
	remoteAddr  string
	result      parserInfo
}{
	{
		description: "basic ipv4 request",
		url:         "/provision/00:00:00:ff:ff:ff",
		remoteAddr:  "10.5.1.1:9873",
		result: parserInfo{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "ipxe",
		},
	},
	{
		description: "basic ipv6 request",
		url:         "/provision/00:00:00:ff:ff:ff",
		remoteAddr:  "[fd00:5::1:1]:9873",
		result: parserInfo{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "fd00:5::1:1",
			remoteport: 9873,
			stage:      "ipxe",
		},
	},
}

func Test_ParseReq(t *testing.T) {
	for _, tt := range parseReqTests {
		t.Run(tt.description, func(t *testing.T) {
			req := &http.Request{
				URL:        &url.URL{Path: tt.url},
				RemoteAddr: tt.remoteAddr,
			}
			result, err := parseReq(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}
