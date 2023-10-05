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
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	overlayDir := fmt.Sprintf("%s/overlay", tmpdir)
	err = os.MkdirAll(overlayDir, 0o755)
	assert.NoError(t, err)
	importDir := fmt.Sprintf("%s/test", overlayDir)
	err = os.MkdirAll(importDir, 0o755)
	assert.NoError(t, err)
	importDir2 := fmt.Sprintf("%s/test2", overlayDir)
	err = os.MkdirAll(importDir2, 0o755)
	assert.NoError(t, err)

	file, err := os.CreateTemp(tmpdir, "file")
	assert.NoError(t, err)
	file.Close()
	err = os.Chmod(file.Name(), 0o755)
	assert.NoError(t, err)

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
			t.Errorf("Should receive error when running command")
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
	t.Run("wwctl overlay import test with changed permissions", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"--mode", "644", "-n", "test", file.Name()})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		err = baseCmd.Execute()
		assert.NoError(t, err)
		assert.FileExists(t, importDir2+file.Name())
	})
}
