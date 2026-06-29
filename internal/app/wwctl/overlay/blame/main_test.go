package blame

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

const testNodesConf = `nodeprofiles:
  default:
    system overlay:
      - profile-system
    runtime overlay:
      - profile-runtime
nodes:
  node1:
    profiles:
      - default
    system overlay:
      - node-system
    runtime overlay:
      - node-runtime
`

func TestOverlayBlame(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	createTestOverlayRoots(env)
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/profile.conf", "profile\n")
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/template.conf.ww", "template\n")
	env.Symlink("/target", "var/lib/warewulf/overlays/profile-system/rootfs/etc/profile.link")
	env.WriteFile("var/lib/warewulf/overlays/node-system/rootfs/etc/profile.conf", "node\n")
	env.WriteFile("var/lib/warewulf/overlays/node-system/rootfs/etc/systemd/system/custom.service", "service\n")
	env.WriteFile("var/lib/warewulf/overlays/profile-runtime/rootfs/run/runtime.conf", "runtime\n")
	env.WriteFile("var/lib/warewulf/overlays/node-runtime/rootfs/etc/runtime-node.conf", "node-runtime\n")

	stdout, err := executeCommand("node1")
	assert.NoError(t, err)
	assert.Equal(t, `/etc/profile.conf                   profile-system   [system overlay]
/etc/profile.link                   profile-system   [system overlay]
/etc/template.conf                  profile-system   [system overlay]
/etc/profile.conf                   node-system      [system overlay]
/etc/systemd/system/custom.service  node-system      [system overlay]
/run/runtime.conf                   profile-runtime  [runtime overlay]
/etc/runtime-node.conf              node-runtime     [runtime overlay]
`, stdout)
}

func TestOverlayBlameJSONOutput(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	createTestOverlayRoots(env)
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/profile.conf", "profile\n")
	env.WriteFile("var/lib/warewulf/overlays/profile-runtime/rootfs/run/runtime.conf", "runtime\n")

	stdout, err := executeCommand("--format", "json", "node1")
	assert.NoError(t, err)
	assert.NotContains(t, stdout, "[system overlay]")

	var payload []blameLine
	if !assert.NoError(t, json.Unmarshal([]byte(stdout), &payload)) {
		return
	}
	assert.Equal(t, []blameLine{
		{Path: "/etc/profile.conf", Overlay: "profile-system", Context: "system"},
		{Path: "/run/runtime.conf", Overlay: "profile-runtime", Context: "runtime"},
	}, payload)
}

func TestOverlayBlameInvalidFormat(t *testing.T) {
	_, err := executeCommand("--format", "yaml", "node1")
	assert.ErrorContains(t, err, `invalid format "yaml": expected table or json`)
}

func TestOverlayBlameShowModeChangesIncludesDirectories(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	createTestOverlayRoots(env)
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/systemd/system/sshd.service", "service\n")

	stdout, err := executeCommand("--show-mode-changes", "node1")
	assert.NoError(t, err)
	assert.Equal(t, `/etc                              profile-system  [system overlay]
/etc/systemd                      profile-system  [system overlay]
/etc/systemd/system               profile-system  [system overlay]
/etc/systemd/system/sshd.service  profile-system  [system overlay]
`, stdout)
}

func TestOverlayBlameShowModeChangesKeepsWWDirectorySuffix(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	createTestOverlayRoots(env)
	env.MkdirAll("var/lib/warewulf/overlays/profile-system/rootfs/foo.ww")

	stdout, err := executeCommand("--show-mode-changes", "node1")
	assert.NoError(t, err)
	assert.Equal(t, "/foo.ww  profile-system  [system overlay]\n", stdout)
}

func TestOverlayBlamePathPrefix(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	createTestOverlayRoots(env)
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/app.conf", "app\n")
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc2/app.conf", "app\n")
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/opt/app.conf", "app\n")

	stdout, err := executeCommand("--path-prefix", "/etc", "node1")
	assert.NoError(t, err)
	assert.Equal(t, "/etc/app.conf  profile-system  [system overlay]\n", stdout)
}

func TestOverlayBlameTemplateGeneratedPaths(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf", testNodesConf)
	createTestOverlayRoots(env)
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/source.conf.ww", `{{ file "multi-a.conf" }}
A
{{ file "multi-b.conf" }}
B
`)
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/link-from-template.ww", `{{ softlink "/target" }}
`)
	env.WriteFile("var/lib/warewulf/overlays/profile-system/rootfs/etc/abort.conf.ww", `{{ abort }}
`)

	stdout, err := executeCommand("node1")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "/etc/multi-a.conf")
	assert.Contains(t, stdout, "/etc/multi-b.conf")
	assert.Contains(t, stdout, "/etc/link-from-template")
	assert.NotContains(t, stdout, "/etc/source.conf")
	assert.NotContains(t, stdout, "/etc/abort.conf")
}

func TestOverlayBlameRequiresNode(t *testing.T) {
	_, err := executeCommand()
	assert.Error(t, err)
}

func TestOverlayBlameRejectsInvalidOverlayName(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	env.WriteFile("etc/warewulf/nodes.conf", `nodeprofiles:
  default: {}
nodes:
  node1:
    system overlay:
      - ../outside
`)

	_, err := executeCommand("node1")
	assert.ErrorContains(t, err, "overlay names contains illegal characters")
}

func executeCommand(args ...string) (string, error) {
	cmd := GetCommand()
	cmd.SetArgs(args)
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)
	cmd.SetErr(new(bytes.Buffer))
	err := cmd.Execute()
	return stdout.String(), err
}

func createTestOverlayRoots(env *testenv.TestEnv) {
	for _, overlayName := range []string{"profile-system", "node-system", "profile-runtime", "node-runtime"} {
		env.MkdirAll("var/lib/warewulf/overlays/" + overlayName + "/rootfs")
	}
}
