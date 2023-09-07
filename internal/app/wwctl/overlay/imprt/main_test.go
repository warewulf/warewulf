package imprt

import (
	"fmt"
	"os"
	"testing"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/stretchr/testify/assert"
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

	inDb := `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes: {}
`
	conf_yml := `
WW_INTERNAL: 0
    `

	conf := warewulfconf.New()
	err = conf.Parse([]byte(conf_yml))
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

		baseCmd.SetArgs([]string{"-p", "-n", "test", file.Name()})
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
