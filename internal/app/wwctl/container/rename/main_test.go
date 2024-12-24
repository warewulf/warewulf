package rename

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	containerList "github.com/warewulf/warewulf/internal/app/wwctl/container/list"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Rename(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile(path.Join(testenv.WWChrootdir, "test-container/rootfs/file"), `test`)
	defer env.RemoveAll()
	warewulfd.SetNoDaemon()

	// first we will verify that there is an existing container
	t.Run("container list", func(t *testing.T) {
		verifyContainerListOutput(t, "test-container")
	})

	// then rename it
	t.Run("container rename", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetOut(os.Stdout)
		baseCmd.SetErr(os.Stdout)
		baseCmd.SetArgs([]string{"test-container", "test-container-rename"})
		err := baseCmd.Execute()
		assert.NoError(t, err)
	})

	// retrieve again
	t.Run("Container list", func(t *testing.T) {
		verifyContainerListOutput(t, "test-container-rename")
	})
}

func verifyContainerListOutput(t *testing.T, content string) {
	baseCmd := containerList.GetCommand()
	buf := new(bytes.Buffer)
	baseCmd.SetOut(buf)
	baseCmd.SetErr(buf)
	wwlog.SetLogWriterErr(buf)
	wwlog.SetLogWriterInfo(buf)
	err := baseCmd.Execute()
	assert.NoError(t, err)

	assert.NotEmpty(t, buf.String(), "output should not be empty")
	assert.Contains(t, buf.String(), content)
}
