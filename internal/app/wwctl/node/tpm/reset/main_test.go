package reset

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

func Test_Reset(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()

	tests := []struct {
		name     string
		args     []string
		tpmFiles map[string]string
		stdout   string
		expectRemoved bool
	}{
		{
			name: "node with no tpm data",
			args: []string{"node1"},
			stdout: `
No TPM quote found for node node1
`,
		},
		{
			name: "node with only current quote",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `{"current": {"quote": "current_quote", "signature": "sig", "ak_pub": "ak", "nonce": "nonce"}}`,
			},
			stdout: `
Removed Current TPM quote for node node1
`,
			expectRemoved: true,
		},
		{
			name: "node with new quote",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `{"current": {"quote": "current_quote", "signature": "sig", "ak_pub": "ak", "nonce": "nonce"}, "new": {"quote": "new_quote", "signature": "sig", "ak_pub": "ak", "nonce": "nonce"}}`,
			},
			stdout: `
Moved NEW TPM quote to Current for node node1
`,
			expectRemoved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env.RemoveAll()
			env = testenv.New(t)
			_ = env.Configure()

			for nodeName, content := range tt.tpmFiles {
				tpmPath := path.Join("srv/warewulf/overlays", nodeName, "tpm.json")
				env.WriteFile(tpmPath, content)
			}

			buf := new(bytes.Buffer)
			wwlog.SetLogWriter(buf)
			wwlog.SetLogWriterErr(buf)
			wwlog.SetLogWriterInfo(buf)

			baseCmd := GetCommand()
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			baseCmd.SetArgs(tt.args)

			err := baseCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			expected := strings.TrimSpace(tt.stdout)

			assert.Equal(t, expected, strings.TrimSpace(output))

			// Check file system state
			if tt.expectRemoved {
				tpmPath := path.Join(env.BaseDir, "srv/warewulf/overlays/node1/tpm.json")
				_, err := os.Stat(tpmPath)
				assert.True(t, os.IsNotExist(err), "File should have been removed")
			}
		})
	}
}
