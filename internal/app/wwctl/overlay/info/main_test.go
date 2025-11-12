package info

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Overlay_Variables(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	wwlog.SetLogLevel(wwlog.DEBUG)
	warewulfd.SetNoDaemon()

	templateContent := `
{{/* .Kernel.Tags.foo: "some help text" */}}
{{/* wwdoc1: First Line */}}
{{/* wwdoc2: Second Line */}}
{{ .Kernel.Tags.foo }}
{{ .Node.Tags.bar }}
{{ .Cluster.Tags.baz }}
{{ .Kernel.Vars }}
`
	env.WriteFile("var/lib/warewulf/overlays/test-overlay/test.ww", templateContent)

	t.Run("overlay variables", func(t *testing.T) {
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)

		baseCmd.SetArgs([]string{"test-overlay", "test.ww"})
		err := baseCmd.Execute()
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "OVERLAY VARIABLE")
		assert.Contains(t, output, ".Kernel.Tags.foo")
		assert.Contains(t, output, "some help text")
		assert.Regexp(t, `(?s)First Line.*Second Line`, output, "First Line should come before Second Line")
	})

	t.Run("overlay variables no file", func(t *testing.T) {
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)

		baseCmd.SetArgs([]string{"test-overlay", "no-file.ww"})
		err := baseCmd.Execute()
		assert.Error(t, err)
	})

	t.Run("overlay variables no overlay", func(t *testing.T) {
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)

		baseCmd.SetArgs([]string{"no-overlay", "test.ww"})
		err := baseCmd.Execute()
		assert.Error(t, err)
	})
}
