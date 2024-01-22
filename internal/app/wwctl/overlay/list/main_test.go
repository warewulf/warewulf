package list

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/hpcng/warewulf/internal/pkg/testenv"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func Test_Overlay_List(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile(t, path.Join(testenv.WWOverlaydir, "testoverlay/email.ww"), `
{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}
`)
	defer env.RemoveAll(t)
	warewulfd.SetNoDaemon()
	t.Run("overlay list", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		verifyOutput(t, baseCmd, "testoverlay")
	})
	t.Run("overlay list all", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-a"})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		verifyOutput(t, baseCmd, "email.ww")
	})

	t.Run("overlay list all with output yaml", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-a", "-o", "yaml"})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Overlays:\n")
		assert.Contains(t, buf.String(), "Files/Dirs: email.ww\n")
	})

	t.Run("overlay list all with output json", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-a", "-o", "json"})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "{\"Overlays\":{\"testoverlay\":[{\"Files/Dirs\":\"email.ww\"}]}}\n")
	})

	t.Run("overlay list all with output csv", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-a", "-o", "csv"})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		verifyOutput(t, baseCmd, "OVERLAY NAME,FILES/DIRS\ntestoverlay,email.ww\n")
	})

	t.Run("overlay list all with output text", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-a", "-o", "text"})
		baseCmd.SetOut(nil)
		baseCmd.SetErr(nil)
		verifyOutput(t, baseCmd, "  OVERLAY NAME  FILES/DIRS  \n  testoverlay   email.ww    \n")
	})
}

func verifyOutput(t *testing.T, baseCmd *cobra.Command, content string) {
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
