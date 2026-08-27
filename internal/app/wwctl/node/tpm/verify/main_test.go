package verify

import (
	"bytes"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Verify(t *testing.T) {
	warewulfd.SetNoDaemon()
	env := testenv.New(t)
	defer env.RemoveAll()

	tests := []struct {
		name     string
		args     []string
		tpmFiles map[string]string
		stdout   string
	}{
		{
			name: "node with no tpm data",
			args: []string{"node1"},
			stdout: `
WARN   : reading tpm quote for node node1: open BASEDIR/srv/warewulf/overlays/node1/tpm.json: no such file or directory
`,
		},
		{
			name: "node with invalid tpm data",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `invalid json`,
			},
			stdout: `
WARN   : unmarshalling quote for node node1: invalid character 'i' looking for beginning of value
`,
		},
		{
			name: "node with incomplete tpm data",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `{"id":"node1"}`,
			},
			stdout: `
Node: node1
TPM Manufacturer: Unknown
ERROR  : Verifying node node1: TPM Quote not available
`,
		},
		{
			name: "node with garbage eventlog",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `{"id":"node1", "eventlog":"Z2FyYmFnZQ=="}`,
			},
			stdout: `
Node: node1
TPM Manufacturer: Unknown
ERROR  : Verifying node node1: TPM Quote not available
`,
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
			expected = strings.ReplaceAll(expected, "BASEDIR", env.BaseDir)

			assert.Equal(t, expected, strings.TrimSpace(output))
		})
	}
}
