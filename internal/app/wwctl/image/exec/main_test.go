package exec

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func mockChildCmd(cmd *cobra.Command, args []string) error {
	child := exec.Command("/usr/bin/echo", args...)
	child.Stdin = cmd.InOrStdin()
	child.Stdout = cmd.OutOrStdout()
	child.Stderr = cmd.ErrOrStderr()
	return child.Run()
}

func Test_Exec(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.MkdirAll("/var/lib/warewulf/chroots/test/rootfs")
	childCommandFunc = mockChildCmd
	defer func() {
		childCommandFunc = runChildCmd
	}()
	warewulfd.SetNoDaemon()

	tests := []struct {
		name   string
		args   []string
		stdout string
		build  bool
	}{
		{
			name:   "plain test",
			args:   []string{"test", "/bin/true"},
			stdout: `--loglevel 20 image exec __child test -- /bin/true`,
			build:  true,
		},
		{
			name:   "test with --bind",
			args:   []string{"test", "--bind", "/tmp", "/bin/true"},
			stdout: `--loglevel 20 image exec __child test --bind /tmp -- /bin/true`,
			build:  true,
		},
		{
			name:   "test with --node",
			args:   []string{"test", "--node", "node1", "/bin/true"},
			stdout: `--loglevel 20 image exec __child test --node node1 -- /bin/true`,
			build:  true,
		},
		{
			name:   "test with --build=false",
			args:   []string{"test", "--build=false", "/bin/true"},
			stdout: `--loglevel 20 image exec __child test -- /bin/true`,
			build:  false,
		},
		{
			name:   "test with --node and --bind",
			args:   []string{"test", "--bind", "/tmp", "--node", "node1", "/bin/true"},
			stdout: `--loglevel 20 image exec __child test --bind /tmp --node node1 -- /bin/true`,
			build:  true,
		},
		{
			name:   "test with complex command",
			args:   []string{"test", "/bin/bash", "echo 'hello'"},
			stdout: `--loglevel 20 image exec __child test -- /bin/bash echo 'hello'`,
			build:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				binds = []string{}
				nodeName = ""
				Build = true
				SyncUser = false
				os.Remove(env.GetPath("/srv/warewulf/image/test.img"))
				os.Remove(env.GetPath("/srv/warewulf/image/test.img.gz"))
			}()
			cmd := GetCommand()
			cmd.SetArgs(tt.args)
			out := bytes.NewBufferString("")
			err := bytes.NewBufferString("")
			cmd.SetOut(out)
			cmd.SetErr(err)
			assert.NoError(t, cmd.Execute())
			assert.NotEmpty(t, out.String())
			assert.Contains(t, out.String(), tt.stdout)
			if tt.build {
				assert.FileExists(t, env.GetPath("/srv/warewulf/image/test.img"))
				assert.FileExists(t, env.GetPath("/srv/warewulf/image/test.img.gz"))
			} else {
				assert.NoFileExists(t, env.GetPath("/srv/warewulf/image/test.img"))
				assert.NoFileExists(t, env.GetPath("/srv/warewulf/image/test.img.gz"))
			}
		})
	}
}
