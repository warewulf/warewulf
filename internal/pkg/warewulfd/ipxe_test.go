package warewulfd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var ipxeHandlerTests = []struct {
	description string
	url         string
	body        string
	status      int
	ip          string
}{
	{
		"ipxe with NetDevs, KernelVersion, and Authority",
		"/ipxe/00:00:00:00:00:ff",
		"1.1.1 ifname=net:00:00:00:00:00:ff  10.10.0.1 fd00:10::1 10.10.0.1:9873",
		200,
		"10.10.10.12:9873",
	},
	{
		"ipxe over ipv6",
		"/ipxe/00:00:00:00:00:ff",
		"1.1.1 ifname=net:00:00:00:00:00:ff  10.10.0.1 fd00:10::1 [fd00:10::1]:9873",
		200,
		"[fd00:10::10:12]:9873",
	},
}

func Test_HandleIpxe(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("/etc/warewulf/nodes.conf", `nodes:
  n3:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:ff
        device: net
    ipxe template: test
    kernel:
      version: 1.1.1`)

	env.WriteFile("/etc/warewulf/ipxe/test.ipxe", "{{.KernelVersion}}{{range $devname, $netdev := .NetDevs}}{{if and $netdev.Hwaddr $netdev.Device}} ifname={{$netdev.Device}}:{{$netdev.Hwaddr}} {{end}}{{end}} {{.Ipaddr}} {{.Ipaddr6}} {{.Authority}}")

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse
	conf.Ipaddr = "10.10.0.1"
	conf.Ipaddr6 = "fd00:10::1"

	for _, tt := range ipxeHandlerTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleIpxe(w, req)
			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			data, readErr := io.ReadAll(res.Body)
			assert.NoError(t, readErr)
			if tt.body != "" {
				assert.Equal(t, tt.body, string(data))
			}
			assert.Equal(t, tt.status, res.StatusCode)
		})
	}
}
