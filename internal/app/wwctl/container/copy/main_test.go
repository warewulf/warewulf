package copy

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Copy(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile(path.Join(testenv.WWChrootdir, "test-container/rootfs/bin/sh"), `test`)
	defer env.RemoveAll()
	warewulfd.SetNoDaemon()

	t.Run("container copy without build", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"test-container", "test-container-copy-without-build"})
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.NoFileExists(t, path.Join(env.BaseDir, testenv.WWProvisiondir, "container", "test-container-copy-without-build.img"))
	})

	t.Run("container copy with build", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-b", "test-container", "test-container-copy"})
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.FileExists(t, path.Join(env.BaseDir, testenv.WWProvisiondir, "container", "test-container-copy.img"))
	})
}
