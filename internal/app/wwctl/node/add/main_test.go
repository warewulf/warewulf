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
	t.Helper()
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
	buf := new(bytes.Buffer)
	baseCmd.SetOut(buf)
	baseCmd.SetErr(buf)
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		stdout  string
		outDb   string
	}{
		{name: "single node add",
			args:    []string{"n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    network devices:
      default: {}
`},
		{name: "double node add",
			args:    []string{"n0[1-2]"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    network devices:
      default: {}
  n02:
    profiles:
    - default
    network devices:
      default: {}
`},
	}
	for _, tt := range tests {
		db, err = node.TestNew([]byte(nodes_yml))
		assert.NoError(t, err)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd.SetArgs(tt.args)
			err = baseCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Got unwanted error: %s", err)
				return
			}
			dump := string(db.DBDump())
			if dump != tt.outDb {
				t.Errorf("DB dump is wrong, got:'%s'\nwant:'%s'", dump, tt.outDb)
				return
			}
			if buf.String() != tt.stdout {
				t.Errorf("Got wrong output, got:'%s'\nwant:'%s'", buf.String(), tt.stdout)
			}
		})
	}
}
