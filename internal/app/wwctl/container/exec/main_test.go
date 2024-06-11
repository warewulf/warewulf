package exec

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func mockChildCmd(cmd *cobra.Command, args []string) error {
	child := exec.Command("/usr/bin/echo", args...)
	child.Stdin = cmd.InOrStdin()
	child.Stdout = cmd.OutOrStdout()
	child.Stderr = cmd.ErrOrStderr()
	return child.Run()
}

func Test_Exec_with_bind(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.MkdirAll(t, "var/lib/warewulf/chroots/alpine/rootfs")
	childCommandFunc = mockChildCmd
	defer func() {
		childCommandFunc = runChildCmd
	}()

	cmd := GetCommand()
	out := bytes.NewBufferString("")
	err := bytes.NewBufferString("")
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.SetArgs([]string{"--bind", "/mnt:/mnt", "alpine", "--", "/bin/echo", "Hello, world!"})
	assert.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "alpine")
	assert.Contains(t, out.String(), "/bin/echo Hello, world!")
	assert.Contains(t, out.String(), "--bind /mnt:/mnt")
	assert.Empty(t, err.String())
}
