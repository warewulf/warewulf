package api

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

var imageTests = map[string]struct {
	initFiles         []string
	request           func(serverURL string) (*http.Request, error)
	response          string
	status            int
	resultFiles       []string
	resultAbsentFiles []string
	authenticate      bool
}{
	"test no authentication": {
		initFiles: []string{
			"/var/lib/warewulf/chroots/test-image/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/images", nil)
		},
		response:     fmt.Sprintln("Unauthorized"),
		status:       http.StatusUnauthorized,
		authenticate: false,
	},

	"test get all images": {
		initFiles: []string{
			"/var/lib/warewulf/chroots/test-image/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/images", nil)
		},
		response:     `{"test-image": {"kernels":[], "size":0, "buildtime":0, "writable":true}}`,
		authenticate: true,
	},

	"test get single image": {
		initFiles: []string{
			"/var/lib/warewulf/chroots/test-image/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, serverURL+"/api/images/test-image", nil)
		},
		response:     `{"kernels":[], "size":0, "buildtime":0, "writable":true}`,
		authenticate: true,
	},

	"test build image": {
		initFiles: []string{
			"/var/lib/warewulf/chroots/test-image/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPost, serverURL+"/api/images/test-image/build?force=true&default=true", nil)
		},
		response: `{"kernels":[], "size":512, "buildtime":"<<PRESENCE>>", "writable":true}`,
		resultFiles: []string{
			"/srv/warewulf/images/test-image.img",
			"/srv/warewulf/images/test-image.img.gz",
		},
		authenticate: true,
	},

	"test rename image": {
		initFiles: []string{
			"/var/lib/warewulf/chroots/test-image/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodPatch, serverURL+"/api/images/test-image?build=true", bytes.NewBuffer([]byte(`{"name": "new-image"}`)))
		},
		response:     `{"kernels":[], "size":512, "buildtime":"<<PRESENCE>>", "writable":true}`,
		authenticate: true,
	},

	"test delete image": {
		initFiles: []string{
			"/var/lib/warewulf/chroots/new-image/rootfs/file",
		},
		request: func(serverURL string) (*http.Request, error) {
			return http.NewRequest(http.MethodDelete, serverURL+"/api/images/new-image", nil)
		},
		response: `{"kernels":[], "size":0, "buildtime":"<<PRESENCE>>", "writable":true}`,
		resultAbsentFiles: []string{
			"/var/lib/warewulf/chroots/new-image",
			"/srv/warewulf/images/new-image.img",
			"/srv/warewulf/images/new-image.img.gz",
		},
		authenticate: true,
	},
}

func TestImageAPI(t *testing.T) {
	authData := `
users:
- name: admin
  password hash: $2b$05$5QVWDpiWE7L4SDL9CYdi3O/l6HnbNOLoXgY2sa1bQQ7aSBKdSqvsC
`

	for name, tt := range imageTests {
		t.Run(name, func(t *testing.T) {
			warewulfd.SetNoDaemon()
			env := testenv.New(t)
			defer env.RemoveAll()

			// Create test files
			for _, fileName := range tt.initFiles {
				env.CreateFile(fileName)
			}

			auth := config.NewAuthentication()
			err := auth.ParseFromRaw([]byte(authData))
			assert.NoError(t, err)

			allowedNets := []net.IPNet{
				{
					IP:   net.IPv4(127, 0, 0, 0),
					Mask: net.CIDRMask(8, 32),
				},
			}
			srv := httptest.NewServer(Handler(auth, allowedNets))
			defer srv.Close()

			req, err := tt.request(srv.URL)
			assert.NoError(t, err)

			if tt.authenticate {
				req.SetBasicAuth("admin", "admin")
			}

			resp, err := http.DefaultTransport.RoundTrip(req)
			assert.NoError(t, err)

			expectedStatus := tt.status
			if expectedStatus == 0 {
				expectedStatus = http.StatusOK
			}
			assert.Equal(t, expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.NoError(t, resp.Body.Close())

			if expectedStatus == http.StatusUnauthorized {
				// For plain text responses like Unauthorized
				assert.Equal(t, tt.response, string(body))
			} else {
				// For JSON responses
				ja := jsonassert.New(t)
				ja.Assertf(string(body), tt.response) //nolint:govet
			}

			for _, fileName := range tt.resultFiles {
				assert.FileExists(t, env.GetPath(fileName))
			}

			for _, fileName := range tt.resultAbsentFiles {
				assert.NoFileExists(t, env.GetPath(fileName))
			}
		})
	}
}
