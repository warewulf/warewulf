package add

import (
	"bytes"
	"testing"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/stretchr/testify/assert"
)

func Test_Add(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		stdout  string
		outDb   string
	}{
		{
			name:    "single profile add",
			args:    []string{"--yes", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  p01:
    network devices:
      default: {}
nodes: {}
`,
		},
		{
			name:    "single profile add with netname and netdev",
			args:    []string{"--yes", "--netname", "primary", "--netdev", "eno3", "p02"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  p02:
    network devices:
      primary:
        device: eno3
nodes: {}
`,
		},
	}
	conf_yml := `
WW_INTERNAL: 0
    `
	nodes_yml := `
WW_INTERNAL: 43
`
	conf := warewulfconf.New()
	err := conf.Read([]byte(conf_yml))
	assert.NoError(t, err)
	db, err := node.TestNew([]byte(nodes_yml))
	assert.NoError(t, err)
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		db, err = node.TestNew([]byte(nodes_yml))
		assert.NoError(t, err)
		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err = baseCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Got unwanted error: %s", err)
				t.FailNow()
			}
			dump := string(db.DBDump())
			if dump != tt.outDb {
				t.Errorf("DB dump is wrong, got:'%s'\nwant:'%s'", dump, tt.outDb)
				t.FailNow()
			}
			if buf.String() != tt.stdout {
				t.Errorf("Got wrong output, got:'%s'\nwant:'%s'", buf.String(), tt.stdout)
				t.FailNow()
			}
		})
	}
}
