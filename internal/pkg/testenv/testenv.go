// Package testenv provides functions and data structures for
// constructing and manipulating a temporary Warewulf environment for
// use during automated testing.
//
// The testenv package should only be used in tests.
package testenv

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/stretchr/testify/assert"
)

const initWarewulfConf = `WW_INTERNAL: 0`
const initDefaultsConf = `WW_INTERNAL: 43`
const initNodesConf = `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  node1: {}
`

type TestEnv struct {
	BaseDir string
}

// New creates a test environment in a temporary directory and configures
// Warewulf to use it.
//
// Caller is responsible to delete env.BaseDir by calling
// env.RemoveAll. Note that this does not restore Warewulf to its
// previous state.
//
// Asserts no errors occur.
func New(t *testing.T) (env *TestEnv) {
	env = new(TestEnv)

	tmpDir, err := os.MkdirTemp(os.TempDir(), "ww4test-*")
	assert.NoError(t, err)
	env.BaseDir = tmpDir

	env.WriteFile(t, "etc/warewulf/nodes.conf", initNodesConf)
	env.WriteFile(t, "etc/warewulf/warewulf.conf", initWarewulfConf)
	env.WriteFile(t, "share/warewulf/defaults.conf", initDefaultsConf)

	// re-read warewulf.conf
	conf := config.New()
	err = conf.Read(env.GetPath("etc/warewulf/warewulf.conf"))
	assert.NoError(t, err)

	conf.Paths.Sysconfdir = env.GetPath("etc")
	conf.Paths.Bindir = env.GetPath("bin")
	conf.Paths.Datadir = env.GetPath("share")
	conf.Paths.Localstatedir = env.GetPath("var/local")
	conf.Paths.Srvdir = env.GetPath("srv")
	conf.Paths.Tftpdir = env.GetPath("srv/tftp")
	conf.Paths.Firewallddir = env.GetPath("usr/lib/firewalld/services")
	conf.Paths.Systemddir = env.GetPath("usr/lib/systemd/system")
	conf.Paths.WWOverlaydir = env.GetPath(path.Join(conf.Paths.Localstatedir, "warewulfoverlays"))
	conf.Paths.WWChrootdir = env.GetPath(path.Join(conf.Paths.Localstatedir, "warewulf/chroots"))
	conf.Paths.WWProvisiondir = env.GetPath(path.Join(conf.Paths.Srvdir, "warewulf"))
	conf.Paths.WWClientdir = env.GetPath("warewulf")

	for _, confPath := range []string{
		conf.Paths.Sysconfdir,
		conf.Paths.Bindir,
		conf.Paths.Datadir,
		conf.Paths.Localstatedir,
		conf.Paths.Srvdir,
		conf.Paths.Tftpdir,
		conf.Paths.Firewallddir,
		conf.Paths.Systemddir,
		conf.Paths.WWOverlaydir,
		conf.Paths.WWChrootdir,
		conf.Paths.WWProvisiondir,
		conf.Paths.WWClientdir,
	} {
		env.MkdirAll(t, confPath)
	}

	// node.init() has already run, so set the config path again
	node.ConfigFile = env.GetPath("etc/warewulf/nodes.conf")

	return
}

// GetPath returns the absolute path name for fileName specified
// relative to the test environment.
func (env *TestEnv) GetPath(fileName string) string {
	return path.Join(env.BaseDir, fileName)
}

// MkdirAll creates dirName and any intermediate directories relative
// to the test environment.
//
// Asserts no errors occur.
func (env *TestEnv) MkdirAll(t *testing.T, dirName string) {
	err := os.MkdirAll(env.GetPath(dirName), 0755)
	assert.NoError(t, err)
}

// WriteFile writes content to fileName, creating any necessary
// intermediate directories relative to the test environment.
//
// Asserts no errors occur.
func (env *TestEnv) WriteFile(t *testing.T, fileName string, content string) {
	dirName := filepath.Dir(fileName)
	env.MkdirAll(t, dirName)

	f, err := os.Create(env.GetPath(fileName))
	assert.NoError(t, err)
	defer f.Close()
	_, err = f.WriteString(content)
	assert.NoError(t, err)
}

// WriteFileAbs uses an absloute path in opposite to WriteFile
func (env *TestEnv) WriteFileAbs(t *testing.T, fileName string, content string) {
	dirName := filepath.Dir(fileName)
	err := os.MkdirAll(dirName, 0755)
	assert.NoError(t, err)
	f, err := os.Create(fileName)
	assert.NoError(t, err)
	defer f.Close()
	_, err = f.WriteString(content)
	assert.NoError(t, err)
}

// ReadFile returns the content of fileName as converted to a
// string.
//
// Asserts no errors occur.
func (env *TestEnv) ReadFile(t *testing.T, fileName string) string {
	buffer, err := os.ReadFile(env.GetPath(fileName))
	assert.NoError(t, err)
	return string(buffer)
}

// RemoveAll deletes the temporary directory, and all its contents,
// for the test environment.
//
// Asserts no errors occur.
func (env *TestEnv) RemoveAll(t *testing.T) {
	err := os.RemoveAll(env.BaseDir)
	assert.NoError(t, err)
}
