package list

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Overlay_List(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile("var/lib/warewulf/overlays/testoverlay/email.ww", `
{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}
`)
	defer env.RemoveAll()
	warewulfd.SetNoDaemon()
	t.Run("overlay list", func(t *testing.T) {
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "testoverlay")
	})
	t.Run("overlay list all", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"-a"})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "email.ww")
	})
	t.Run("overlay list long", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"--long"})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "email.ww")
	})

	t.Run("overlay list path", func(t *testing.T) {
		baseCmd := GetCommand()
		baseCmd.SetArgs([]string{"--path"})
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), env.BaseDir)
	})
}
