package imprt

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_List(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "warewulf")
	if err != nil {
		t.Errorf("Could not create temp folder: %v", err)
		t.FailNow()
	}
	defer os.RemoveAll(tmpdir)

	overlayDir := fmt.Sprintf("%s/overlay", tmpdir)
	err = os.MkdirAll(overlayDir, 0o755)
	if err != nil {
		t.Errorf("Could not create target folder: %s, err: %v", overlayDir, err)
		t.FailNow()
	}

	importDir := fmt.Sprintf("%s/test", overlayDir)
	err = os.MkdirAll(importDir, 0o755)
	if err != nil {
		t.Errorf("Could not create target folder: %s, err: %v", importDir, err)
		t.FailNow()
	}

	file, err := os.CreateTemp(tmpdir, "file")
	if err != nil {
		t.Errorf("Could not create tempfile")
		t.FailNow()
	}
	file.Close()
	err = os.Chmod(file.Name(), 0o755)
	if err != nil {
		t.Errorf("Could not change the file %s mode: %v", file.Name(), err)
		t.FailNow()
	}

	inDb := `nodeprofiles:
  default: {}
nodes: {}
`
	conf_yml := ``

	conf := warewulfconf.New()
	err = conf.Parse([]byte(conf_yml), false)
	assert.NoError(t, err)
	warewulfd.SetNoDaemon()
	conf.Paths.WWOverlaydir = overlayDir

	_, err = node.Parse([]byte(inDb))
	assert.NoError(t, err)
	t.Logf("Running test: wwctl overlay import test\n")
	t.Run("wwctl overlay import test", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-n", "test", file.Name()})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		err = baseCmd.Execute()
		if err == nil {
			t.Errorf("Should recieve error when running command")
			t.FailNow()
		}
		if _, err = os.Stat(importDir + file.Name()); err == nil {
			t.Errorf("Target file %s should not exist", importDir+file.Name())
			t.FailNow()
		}

		baseCmd.SetArgs([]string{"-p", "test", file.Name()})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		err = baseCmd.Execute()
		if err != nil {
			t.Errorf("Received error when running command, err: %v\n", err)
			t.FailNow()
		}
		if _, err = os.Stat(importDir + file.Name()); os.IsNotExist(err) {
			t.Errorf("Target file %s should exist", importDir+file.Name())
			t.FailNow()
		}
	})
}

func Test_Import(t *testing.T) {
	tests := map[string]struct {
		initFiles   []string
		initDirs    []string
		args        []string
		errExpected bool
	}{
		"import a file": {
			initFiles: []string{"importfile"},
			initDirs:  []string{"/var/lib/warewulf/overlays/to1/rootfs"},
			args:      []string{"to1", "importfile"},
		},

		"import missing parent": {
			initFiles:   []string{"importfile"},
			initDirs:    []string{"/var/lib/warewulf/overlays/to1/rootfs"},
			args:        []string{"to1", "importfile", "a/b/importfile"},
			errExpected: true,
		},

		"import create parents": {
			initFiles:   []string{"importfile"},
			initDirs:    []string{"/var/lib/warewulf/overlays/to1/rootfs"},
			args:        []string{"to1", "importfile", "a/b/importfile", "--parents"},
			errExpected: false,
		},

		"import fail overwrite": {
			initFiles:   []string{"importfile", "/var/lib/warewulf/overlays/to1/rootfs/importfile"},
			args:        []string{"to1", "importfile"},
			errExpected: true,
		},

		"import overwrite": {
			initFiles:   []string{"importfile", "/var/lib/warewulf/overlays/to1/rootfs/importfile"},
			args:        []string{"to1", "importfile", "--overwrite"},
			errExpected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			{
				wd, err := os.Getwd()
				assert.NoError(t, err)
				defer func() { assert.NoError(t, os.Chdir(wd)) }()
			}
			assert.NoError(t, os.Chdir(env.GetPath(".")))

			OverwriteFile = false
			CreateDirs = false

			for _, file := range tt.initFiles {
				env.CreateFile(file)
			}
			for _, dir := range tt.initDirs {
				env.MkdirAll(dir)
			}

			cmd := GetCommand()
			cmd.SetArgs(tt.args)
			stdout := new(bytes.Buffer)
			cmd.SetOut(stdout)
			stderr := new(bytes.Buffer)
			cmd.SetErr(stderr)
			err := cmd.Execute()
			if tt.errExpected {
				assert.Error(t, err, stdout)
			} else {
				assert.NoError(t, err, stderr)
			}
		})
	}
}
