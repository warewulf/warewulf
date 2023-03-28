package list

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/stretchr/testify/assert"
)

func Test_List(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		stdout  string
		inDb    string
	}{
		{name: "single node list",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK  
  n01        default            
`,
			inDb: `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`},
	}
	conf_yml := `
WW_INTERNAL: 0
    `
	conf := warewulfconf.New()
	err := conf.Read([]byte(conf_yml))
	assert.NoError(t, err)
	assert.NoError(t, err)
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		_, err = node.TestNew([]byte(tt.inDb))
		assert.NoError(t, err)
		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			old := os.Stdout // keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			err = baseCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Got unwanted error: %s", err)
				t.FailNow()
			}
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				outC <- buf.String()
			}()
			// back to normal state
			w.Close()
			os.Stdout = old // restoring the real stdout
			out := <-outC
			if out != tt.stdout {
				t.Errorf("Got wrong output, got:'%s'\nwant:'%s'", out, tt.stdout)
				t.FailNow()
			}
		})
	}
}
