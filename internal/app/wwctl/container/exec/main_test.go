package exec

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"path"
	"strings"

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
	defer env.RemoveAll(t)
	env.MkdirAll(t, path.Join(testenv.WWChrootdir, "test/rootfs"))
	childCommandFunc = mockChildCmd
	defer func() {
		childCommandFunc = runChildCmd
	}()
	warewulfd.SetNoDaemon()

	tests := []struct {
		name   string
		args   []string
		stdout string
	}{
		{
			name:   "plain test",
			args:   []string{"test", "/bin/true"},
			stdout: `--loglevel 20 container exec __child test -- /bin/true`,
		},
		{
			name:   "test with --bind",
			args:   []string{"test", "--bind", "/tmp", "/bin/true"},
			stdout: `--loglevel 20 container exec __child test --bind /tmp -- /bin/true`,
		},
		{
			name:   "test with --node",
			args:   []string{"test", "--node", "node1", "/bin/true"},
			stdout: `--loglevel 20 container exec __child test --node node1 -- /bin/true`,
		},
		{
			name:   "test with --node and --bind",
			args:   []string{"test", "--bind", "/tmp", "--node", "node1", "/bin/true"},
			stdout: `--loglevel 20 container exec __child test --bind /tmp --node node1 -- /bin/true`,
		},
		{
			name:   "test with complex command",
			args:   []string{"test", "/bin/bash", "echo 'hello'"},
			stdout: `--loglevel 20 container exec __child test -- /bin/bash echo 'hello'`,
		},
	}

	for _, tt := range tests {
		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				binds = []string{}
				nodeName = ""
			}()
			cmd := GetCommand()
			cmd.SetArgs(tt.args)
			out := bytes.NewBufferString("")
			err := bytes.NewBufferString("")
			cmd.SetOut(out)
			cmd.SetErr(err)
			if err := cmd.Execute(); err != nil {
				t.Errorf("Received error when running command, err: %v", err)
				t.FailNow()
			}
			assert.NotEmpty(t, out.String(), "os.stdout should not be empty")
			if !strings.Contains(out.String(), tt.stdout) {
				t.Errorf("Got wrong output, got:\n '%s'\n, but want:\n '%s'\n", out.String(), tt.stdout)
				t.FailNow()
			}
		})
	}
}
