package show

import (
	"bytes"
	"path"
	"testing"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/stretchr/testify/assert"

	"github.com/hpcng/warewulf/internal/pkg/testenv"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
)

var (
	overlayCont = `
{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}
`
)

func Test_Overlay_List(t *testing.T) {
	env := testenv.New(t)
	wwconf := warewulfconf.Get()
	env.WriteFile(t, "etc/warewulf/nodes.conf",
		`WW_INTERNAL: 43
nodeprofiles:
  default:
    tags:
      email: admin@localhost
  empty: {}
nodes:
  node1:
    tags:
      email: admin@node1
  node2: {}
  node3:
    profiles:
      - empty
`)

	env.WriteFileAbs(t, path.Join(wwconf.Paths.WWOverlaydir, "testoverlay/email.ww"), overlayCont)
	defer env.RemoveAll(t)
	warewulfd.SetNoDaemon()
	t.Run("overlay show raw", func(t *testing.T) {
		baseCmd.SetArgs([]string{"testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), overlayCont)
	})
	t.Run("overlay show rendered node tag", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "node1", "testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "admin@node1")
	})
	t.Run("overlay show rendered profile tag", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "node2", "testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "admin@localhost")
	})
	t.Run("overlay show no tag", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "node3", "testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "noMail")
	})
}
