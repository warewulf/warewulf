package list

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_List_Args(t *testing.T) {
	tests := []struct {
		args   []string
		output string
		fail   bool
	}{
		{args: []string{""},
			output: `  CONTAINER NAME
  test
`,
			fail: false,
		},
		{args: []string{"-ul"},
			output: `  CONTAINER NAME  NODES  KERNEL VERSION  CREATION TIME        MODIFICATION TIME    SIZE
  test            0                      02 Jan 00 03:04 UTC  01 Jan 70 00:00 UTC  0 B
`,
			fail: false,
		},
		{args: []string{"-c"},
			output: `  CONTAINER NAME  NODES  SIZE
  test            0      37 B
`,
			fail: false,
		},
	}
	env := testenv.New(t)
	env.WriteFile(t, path.Join(testenv.WWChrootdir, "test/rootfs/bin/sh"), `This is a fake shell, no pearls here.`)
	// need to touch the files, so that the creation date of the container is constant,
	// modification date of `../chroots/containername` is used as creation date.
	// modification dates of directories change every time a file or subdir is added
	// so we have to make it constant *after* its creation.
	cmd := exec.Command("touch", "-d", "2000-01-02 03:04:05 UTC",
		env.GetPath(path.Join(testenv.WWChrootdir, "test/rootfs")),
		env.GetPath(path.Join(testenv.WWChrootdir, "test")))
	err := cmd.Run()
	assert.NoError(t, err)
	defer env.RemoveAll(t)
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, "_"), func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			stdoutR, stdoutW, _ := os.Pipe()
			os.Stdout = stdoutW
			wwlog.SetLogWriter(os.Stdout)
			baseCmd.SetOut(os.Stdout)
			baseCmd.SetErr(os.Stdout)
			err := baseCmd.Execute()
			if tt.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			stdoutC := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, stdoutR)
				stdoutC <- buf.String()
			}()
			stdoutW.Close()
			stdout := <-stdoutC
			assert.Equal(t, tt.output, stdout)
			assert.Equal(t,
				strings.ReplaceAll(strings.TrimSpace(tt.output), " ", ""),
				strings.ReplaceAll(strings.TrimSpace(stdout), " ", ""))

		})
	}
}
