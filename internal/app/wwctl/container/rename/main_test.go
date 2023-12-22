package rename

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	containerList "github.com/hpcng/warewulf/internal/app/wwctl/container/list"
	"github.com/hpcng/warewulf/internal/pkg/testenv"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/stretchr/testify/assert"
)

func Test_Rename(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile(t, path.Join(testenv.WWChrootdir, "test-container/rootfs/file"), `test`)
	defer env.RemoveAll(t)
	warewulfd.SetNoDaemon()

	// first we will verify that there is an existing container
	t.Run("container list", func(t *testing.T) {
		verifyContainerListOutput(t, "test-container")
	})

	// then rename it
	t.Run("container rename", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
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
	baseCmd.SetOut(nil)
	baseCmd.SetErr(nil)
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	err := baseCmd.Execute()
	assert.NoError(t, err)

	stdoutC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR)
		stdoutC <- buf.String()
	}()
	stdoutW.Close()

	stdout := <-stdoutC
	assert.NotEmpty(t, stdout, "output should not be empty")
	assert.Contains(t, stdout, content)
}
