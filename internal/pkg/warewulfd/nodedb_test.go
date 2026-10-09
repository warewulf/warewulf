package warewulfd

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_GetNodeOrSetDiscoverable(t *testing.T) {
	var tests = map[string]struct {
		nodesConf    string
		hwaddr       string
		node         string
		err          bool
		initFiles    []string
		removedFiles []string
	}{
		"empty": {
			nodesConf: `
nodes: {}
`,
			hwaddr: "00:00:00:00:00:00",
			err:    true,
		},
		"configured": {
			nodesConf: `
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:01
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
		"discoverable": {
			nodesConf: `
nodes:
  n1:
    discoverable: true
    network devices:
      default: {}
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
		"discoverable autobuild": {
			initFiles: []string{
				"/srv/warewulf/overlays/n1/__SYSTEM__.img",
				"/srv/warewulf/overlays/n1/__SYSTEM__.img.gz",
				"/srv/warewulf/overlays/n1/__RUNTIME__.img",
				"/srv/warewulf/overlays/n1/__RUNTIME__.img.gz",
			},
			removedFiles: []string{
				"/srv/warewulf/overlays/n1/__SYSTEM__.img",
				"/srv/warewulf/overlays/n1/__SYSTEM__.img.gz",
				"/srv/warewulf/overlays/n1/__RUNTIME__.img",
				"/srv/warewulf/overlays/n1/__RUNTIME__.img.gz",
			},
			nodesConf: `
nodes:
  n1:
    discoverable: true
    network devices:
      default: {}
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
		"discoverable with primary": {
			nodesConf: `
nodes:
  n1:
    discoverable: true
    primary netdev: default
    network devices:
      default: {}
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
		"discoverable without network": {
			nodesConf: `
nodeprofiles:
  default:
    network devices:
      default:
        netmask: 255.255.255.0
nodes:
  n1:
    profiles:
    - default
    discoverable: true
`,
			hwaddr: "00:00:00:00:00:01",
			node:   "n1",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			for _, file := range tt.initFiles {
				env.CreateFile(file)
			}
			defer env.RemoveAll()
			env.WriteFile("/etc/warewulf/nodes.conf", tt.nodesConf)

			err := LoadNodeDB()
			assert.NoError(t, err)

			node, err := GetOrDiscoverNode(tt.hwaddr, true)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.node, node.Id())
			}
			for _, file := range tt.removedFiles {
				assert.NoFileExists(t, env.GetPath(file), "File should not exist: %s", file)
			}
		})
	}
}

func Test_WarnSingleStageSecureSystem(t *testing.T) {
	nodesConf := `
nodeprofiles:
  twostage:
    tags:
      IPXEMenuEntry: dracut
nodes:
  n1: {}
  n2:
    tags:
      IPXEMenuEntry: dracut_static
  n3:
    profiles:
    - twostage
  n4:
    tags:
      IPXEMenuEntry: initrd
  n5:
    ipxe template: custom
  n6:
    tags:
      GrubMenuEntry: dracut
`
	tests := map[string]struct {
		secure   *warewulfconf.SecureRoutes
		grubBoot bool
		warn     string
	}{
		"system not secured": {secure: warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteRuntime)},
		"default":            {secure: nil},
		"ipxe":               {secure: warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteSystem), warn: "n1, n4, n6 (set tag IPXEMenuEntry=dracut)"},
		"grub":               {secure: warewulfconf.NewSecureRoutes(warewulfconf.SecureRouteSystem), grubBoot: true, warn: "n1, n2, n3, n4, n5 (set tag GrubMenuEntry=dracut)"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("/etc/warewulf/nodes.conf", nodesConf)
			assert.NoError(t, LoadNodeDB())

			conf := warewulfconf.Get()
			conf.Warewulf.SecureP = tt.secure
			conf.Warewulf.GrubBootP = &tt.grubBoot

			buf := new(bytes.Buffer)
			wwlog.SetLogWriter(buf)
			defer wwlog.SetLogWriterErr(os.Stderr)
			defer wwlog.SetLogWriterInfo(os.Stdout)

			warnSingleStageSecureSystem()
			if tt.warn == "" {
				assert.NotContains(t, buf.String(), "do not boot two-stage")
			} else {
				assert.Contains(t, buf.String(), "cannot fetch the system overlay: "+tt.warn)
			}
		})
	}
}

func Test_GetNode(t *testing.T) {
	t.Run("configured node found", func(t *testing.T) {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.WriteFile("/etc/warewulf/nodes.conf", `
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:01
`)
		assert.NoError(t, LoadNodeDB())

		n, err := GetNode("00:00:00:00:00:01")
		assert.NoError(t, err)
		assert.Equal(t, "n1", n.Id())
	})

	t.Run("unknown hwaddr returns error", func(t *testing.T) {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.WriteFile("/etc/warewulf/nodes.conf", `
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:00:00:01
`)
		assert.NoError(t, LoadNodeDB())

		_, err := GetNode("ff:ff:ff:ff:ff:ff")
		assert.Error(t, err)
	})

	t.Run("does not consume discoverable nodes", func(t *testing.T) {
		// This is the key behavioral difference from GetOrDiscoverNode:
		// an unknown hwaddr must NOT cause a discoverable node to be
		// configured, even when one is available.
		env := testenv.New(t)
		defer env.RemoveAll()
		env.WriteFile("/etc/warewulf/nodes.conf", `
nodes:
  n1:
    discoverable: true
    network devices:
      default: {}
`)
		assert.NoError(t, LoadNodeDB())

		conf := warewulfconf.Get()
		nodesConfPath := path.Join(conf.Paths.Sysconfdir, "warewulf", "nodes.conf")
		before, err := os.ReadFile(nodesConfPath)
		assert.NoError(t, err)

		_, err = GetNode("ff:ff:ff:ff:ff:ff")
		assert.Error(t, err)

		// Verify the database was not mutated.
		after, err := os.ReadFile(nodesConfPath)
		assert.NoError(t, err)
		assert.Equal(t, string(before), string(after), "GetNode must not write to nodes.conf")
	})
}
