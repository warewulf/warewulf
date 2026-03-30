package list

import (
	"bytes"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_List(t *testing.T) {
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
NODE   MANUFACTURER  QUOTE  EVENTLOG  GRUB  EKPUB (SHA256)
----   ------------  -----  --------  ----  --------------
node1  --            --     --        --    --
`,
		},
		{
			name: "node with invalid tpm data",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `invalid json`,
			},
			stdout: `
NODE   MANUFACTURER  QUOTE  EVENTLOG  GRUB  EKPUB (SHA256)
----   ------------  -----  --------  ----  --------------
node1  --            ERR    --        --    --
`,
		},
		{
			name: "node with incomplete tpm data (N/A)",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `{"id":"node1"}`,
			},
			stdout: `
NODE   MANUFACTURER  QUOTE  EVENTLOG  GRUB  EKPUB (SHA256)
----   ------------  -----  --------  ----  --------------
node1  Unknown       N/A    N/A       N/A   N/A
`,
		},
		{
			name: "node with garbage eventlog",
			args: []string{"node1"},
			tpmFiles: map[string]string{
				"node1": `{"id":"node1", "eventlog":"Z2FyYmFnZQ=="}`,
			},
			stdout: `
NODE   MANUFACTURER  QUOTE  EVENTLOG  GRUB  EKPUB (SHA256)
----   ------------  -----  --------  ----  --------------
node1  Unknown       N/A    FAIL      FAIL  N/A
`,
		},
		{
			name: "node with key flag but no challenge data",
			args: []string{"node1", "--key"},
			tpmFiles: map[string]string{
				"node1": `{"id":"node1"}`,
			},
			stdout: `
NODE   MANUFACTURER  QUOTE  EVENTLOG  GRUB  SECRET
----   ------------  -----  --------  ----  ------
node1  Unknown       N/A    N/A       N/A   N/A
`,
		},
		{
			name: "node with key flag and challenge data",
			args: []string{"node1", "--key"},
			tpmFiles: map[string]string{
				"node1": `{"id":"node1", "challenge":{"secret":"c2VjcmV0"}}`,
			},
			stdout: `
NODE   MANUFACTURER  QUOTE  EVENTLOG  GRUB  SECRET
----   ------------  -----  --------  ----  ------
node1  Unknown       N/A    N/A       N/A   736563726574
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any previous tpm files
			env.RemoveAll()
			env = testenv.New(t)
			_ = env.Configure()
			keyFlag = false // Reset global flag

			for nodeName, content := range tt.tpmFiles {
				tpmPath := path.Join("srv/warewulf/overlays", nodeName, "tpm.json")
				env.WriteFile(tpmPath, content)
			}

			buf := new(bytes.Buffer)
			baseCmd := GetCommand()
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			baseCmd.SetArgs(tt.args)

			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.stdout), strings.TrimSpace(buf.String()))
		})
	}
}
