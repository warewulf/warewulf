package warewulfd

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

var parseReqTests = []struct {
	description string
	url         string
	rawQuery    string
	remoteAddr  string
	result      parsedRequest
}{
	{
		description: "basic ipv4 request",
		url:         "/provision/00:00:00:ff:ff:ff",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
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
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "fd00:5::1:1",
			remoteport: 9873,
			stage:      "ipxe",
		},
	},
	{
		description: "initramfs dedicated route",
		url:         "/initramfs/00:00:00:ff:ff:ff",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "initramfs",
		},
	},
	{
		description: "grub route with hwaddr in path",
		url:         "/grub/00:00:00:ff:ff:ff",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "grub",
		},
	},
	{
		description: "wwid query param on dedicated route",
		url:         "/kernel/",
		rawQuery:    "wwid=00%3A00%3A00%3Aff%3Aff%3Aff",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "kernel",
		},
	},
	{
		description: "wwid and stage query params on provision route",
		url:         "/provision/",
		rawQuery:    "wwid=00%3A00%3A00%3Aff%3Aff%3Aff&stage=kernel",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "kernel",
		},
	},
	{
		description: "efiboot with wwid and file query params",
		url:         "/efiboot/",
		rawQuery:    "wwid=00%3A00%3A00%3Aff%3Aff%3Aff&file=shim.efi",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "efiboot",
			efifile:    "shim.efi",
		},
	},
	{
		description: "grub with wwid query param",
		url:         "/grub/",
		rawQuery:    "wwid=00%3A00%3A00%3Aff%3Aff%3Aff",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "grub",
		},
	},
	{
		description: "path hwaddr takes priority over wwid query param",
		url:         "/kernel/00:00:00:ff:ff:ff",
		rawQuery:    "wwid=11%3A11%3A11%3A11%3A11%3A11",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "kernel",
		},
	},
	{
		description: "system overlay dedicated route",
		url:         "/system/00:00:00:ff:ff:ff",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "system",
		},
	},
	{
		description: "runtime overlay dedicated route",
		url:         "/runtime/00:00:00:ff:ff:ff",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "runtime",
		},
	},
	{
		description: "stage=grub via provision route serves config (new semantics)",
		url:         "/provision/00:00:00:ff:ff:ff",
		rawQuery:    "stage=grub",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "grub",
		},
	},
	{
		description: "path efifile takes priority over file query param",
		url:         "/efiboot/shim.efi",
		rawQuery:    "wwid=00%3A00%3A00%3Aff%3Aff%3Aff&file=grub.cfg",
		remoteAddr:  "10.5.1.1:9873",
		result: parsedRequest{
			hwaddr:     "00:00:00:ff:ff:ff",
			ipaddr:     "10.5.1.1",
			remoteport: 9873,
			stage:      "efiboot",
			efifile:    "shim.efi",
		},
	},
}

func Test_ParseRequest(t *testing.T) {
	for _, tt := range parseReqTests {
		t.Run(tt.description, func(t *testing.T) {
			req := &http.Request{
				URL:        &url.URL{Path: tt.url, RawQuery: tt.rawQuery},
				RemoteAddr: tt.remoteAddr,
			}
			result, err := parseRequest(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}

func Test_RequiresPrivilegedPort(t *testing.T) {
	tests := map[string]struct {
		secure *warewulfconf.SecureRoutes
		stage  string
		result bool
	}{
		"default runtime": {nil, "runtime", true},
		"default system":  {nil, "system", false},
		"none runtime":    {warewulfconf.NewSecureRoutes(), "runtime", false},
		"system system":   {warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteSystem), "system", true},
		"system runtime":  {warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteSystem), "runtime", false},
		"files system":    {warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteFiles), "system", false},
		"all kernel":      {warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteRuntime, warewulfconf.SecureRouteSystem, warewulfconf.SecureRouteFiles), "kernel", false},
		"runtime runtime": {warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteRuntime), "runtime", true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conf := &warewulfconf.WarewulfConf{SecureP: tt.secure}
			assert.Equal(t, tt.result, requiresPrivilegedPort(tt.stage, conf))
		})
	}
}
