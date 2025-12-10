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
	tests := []struct {
		name           string
		writeFiles     map[string]string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name: "overlay variables",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/test.ww": `
{{/* .Kernel.Tags.foo: "some help text" */}}
{{/* wwdoc1: First Line */}}
{{ .Node.Tags.bar }}
{{/* wwdoc2: Second Line */}}
`,
			},
			args:        []string{"test-overlay", "test.ww"},
			expectError: false,
			expectedOutput: `First Line
Second Line

VARIABLE        OPTION  TYPE    HELP
--------        ------  ----    ----
.Node.Tags.bar          string  
`,
		},
		{
			name: "overlay variables no file",
			writeFiles: map[string]string{
				"var/lib/warewulf/overlays/test-overlay/test.ww": ``,
			},
			args:        []string{"test-overlay", "no-file.ww"},
			expectError: true,
		},
		{
			name:        "overlay variables no overlay",
			args:        []string{"no-overlay", "test.ww"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			warewulfd.SetNoDaemon()

			for path, content := range tt.writeFiles {
				env.WriteFile(path, content)
			}
			baseCmd := GetCommand()
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)

			baseCmd.SetArgs(tt.args)
			err := baseCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedOutput != "" {
				output := buf.String()
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}
