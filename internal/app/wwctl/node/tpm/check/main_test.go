package check

import (
	"bytes"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Check(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()

	tests := []struct {
		name     string
		tpmFiles map[string]string
		stdout   string
		wantErr  bool
	}{
		{
			name: "non-existent file",
			stdout: `
reading quote file: open non-existent: no such file or directory
`,
			wantErr: true,
		},
		{
			name: "invalid json file",
			tpmFiles: map[string]string{
				"invalid.json": `invalid json`,
			},
			stdout: `
unmarshalling quote: invalid character 'i' looking for beginning of value
`,
			wantErr: true,
		},
		{
			name: "incomplete tpm data",
			tpmFiles: map[string]string{
				"incomplete.json": `{"id":"node1"}`,
			},
			stdout: `
File: FILENAME
TPM Manufacturer: Unknown
TPM Quote not available
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env.RemoveAll()
			env = testenv.New(t)

			var args []string
			if tt.name == "non-existent file" {
				args = []string{"non-existent"}
			}

			for fileName, content := range tt.tpmFiles {
				filePath := path.Join(env.BaseDir, fileName)
				err := os.WriteFile(filePath, []byte(content), 0644)
				assert.NoError(t, err)
				args = append(args, filePath)
			}

			buf := new(bytes.Buffer)
			wwlog.SetLogWriter(buf)
			wwlog.SetLogWriterErr(buf)
			wwlog.SetLogWriterInfo(buf)

			baseCmd := GetCommand()
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			baseCmd.SetArgs(args)
			baseCmd.SilenceUsage = true
			baseCmd.SilenceErrors = true

			err := baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := buf.String()
			// If error is returned, append it to output for comparison if it's not already there
			if err != nil && !strings.Contains(output, err.Error()) {
				if output != "" && !strings.HasSuffix(output, "\n") {
					output += "\n"
				}
				output += err.Error()
			}

			expected := strings.TrimSpace(tt.stdout)
			if len(args) > 0 {
				expected = strings.ReplaceAll(expected, "FILENAME", args[0])
			}

			assert.Equal(t, expected, strings.TrimSpace(output))
		})
	}
}
