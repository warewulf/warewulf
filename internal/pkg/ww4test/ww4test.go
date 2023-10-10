package ww4test

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	Env *WarewulfTestEnv
)

func init() {
	Env = new(WarewulfTestEnv)
	Env.WarewulfConf = `WW_INTERNAL: 0`
	Env.DefaultsConf = `WW_INTERNAL: 43`
	Env.NodesConf = `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  node1: {}
`

}

type ConfFile struct {
	// directory in which file exists
	Dir string
	// filename, if empty only directory is create
	FileName string
	// content of file
	Content string
	// permissions of file, 644 for file and 755 for
	// leading directory otherwise
	Mode fs.FileMode
}

type WarewulfTestEnv struct {
	// content of warewulf.conf
	WarewulfConf string
	// path to warewulf.conf
	WarewulfConfFile string
	// content of nodes.conf
	NodesConf string
	// path to nodes.conf
	NodesConfFile string
	// content for defaults.conf
	DefaultsConf string
	// tmpe dir where the files are created
	BaseDir string
	// additional files to be created by New(t)
	WarewulfFiles []ConfFile
}

// creates the given file, if DIR is "" only the dir is created
// any error is returned
func (env *WarewulfTestEnv) CreateFile(file ConfFile) (err error) {
	if env.BaseDir == "" {
		return fmt.Errorf("the variable Basedir is empty, perhaps New() wasn't called")
	}
	var dirM fs.FileMode = file.Mode
	// var fileM fs.FileMode = file.Mode
	if file.Mode == 0 {
		dirM = 0755
		// fileM = 0644
	}
	err = os.MkdirAll(path.Join(env.BaseDir, file.Dir), dirM)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("couldn't create dir: %s", path.Join(env.BaseDir, file.Dir)))
	}
	if file.FileName != "" {
		f, err := os.Create(path.Join(env.BaseDir, file.Dir, file.FileName))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("couldn't open file: %s", path.Join(env.BaseDir, file.Dir, file.FileName)))
		}
		defer f.Close()
		_, err = f.WriteString(file.Content)
		return err
	}
	return
}

/*
Creates a test environment so that warewulf functions can be tested.
nodes.conf content can be provided either via global Env.WarewulfConf.
Caller is responsible to delete the created Env.BaseDir
*/
func (env *WarewulfTestEnv) New(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "ww4test-*")
	assert.NoError(t, err)
	env.BaseDir = tmpDir
	err = env.CreateFile(ConfFile{
		Dir:      "etc/warewulf",
		FileName: "nodes.conf",
		Content:  env.NodesConf,
	})
	assert.NoError(t, err)
	err = env.CreateFile(ConfFile{
		Dir:      "etc/warewulf",
		FileName: "warewulf.conf",
		Content:  env.WarewulfConf,
	})
	assert.NoError(t, err)
	err = env.CreateFile(ConfFile{
		Dir:      "share/warewulf",
		FileName: "defaults.conf",
		Content:  env.WarewulfConf,
	})
	assert.NoError(t, err)
	env.WarewulfConfFile = path.Join(env.BaseDir, "etc/warewulf/warewulf.conf")
	env.NodesConfFile = path.Join(env.BaseDir, "etc/warewulf/nodes.conf")
	conf := config.New()
	err = conf.Read(env.WarewulfConfFile)
	assert.NoError(t, err)
	// init() of node has run before, so set the config path again
	node.ConfigFile = env.NodesConfFile
	conf.Paths.Sysconfdir = path.Join(env.BaseDir, "etc")
	conf.Paths.Bindir = path.Join(env.BaseDir, "bin")
	conf.Paths.Datadir = path.Join(env.BaseDir, "share")
	conf.Paths.Localstatedir = path.Join(env.BaseDir, "var/local")
	conf.Paths.Srvdir = path.Join(env.BaseDir, "srv")
	conf.Paths.Tftpdir = path.Join(env.BaseDir, "srv/tftp")
	conf.Paths.Firewallddir = path.Join(env.BaseDir, "usr/lib/firewalld/services")
	conf.Paths.Systemddir = path.Join(env.BaseDir, "usr/lib/systemd/system")
	conf.Paths.WWOverlaydir = path.Join(env.BaseDir, conf.Paths.Localstatedir, "warewulfoverlays")
	conf.Paths.WWChrootdir = path.Join(env.BaseDir, conf.Paths.Localstatedir, "warewulf/chroots")
	conf.Paths.WWProvisiondir = path.Join(env.BaseDir, conf.Paths.Srvdir, "warewulf")
	conf.Paths.WWClientdir = path.Join(env.BaseDir, "warewulf")
}
