package add

import (
	"bytes"
	"os"
	"testing"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
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
  p01: {}
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

	conf_yml := `WW_INTERNAL: 0`
	tempWarewulfConf, warewulfConfErr := os.CreateTemp("", "warewulf.conf-")
	assert.NoError(t, warewulfConfErr)
	defer os.Remove(tempWarewulfConf.Name())
	_, warewulfConfErr = tempWarewulfConf.Write([]byte(conf_yml))
	assert.NoError(t, warewulfConfErr)
	assert.NoError(t, tempWarewulfConf.Sync())
	warewulfconf.ConfigFile = tempWarewulfConf.Name()

	nodes_yml := `WW_INTERNAL: 43`
	tempNodeConf, nodesConfErr := os.CreateTemp("", "nodes.conf-")
	assert.NoError(t, nodesConfErr)
	defer os.Remove(tempNodeConf.Name())
	node.ConfigFile = tempNodeConf.Name()
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		var err error
		_, err = tempNodeConf.Seek(0, 0)
		assert.NoError(t, err)
		assert.NoError(t, tempNodeConf.Truncate(0))
		_, err = tempNodeConf.Write([]byte(nodes_yml))
		assert.NoError(t, err)
		assert.NoError(t, tempNodeConf.Sync())
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
			config, configErr := node.New()
			assert.NoError(t, configErr)
			dumpBytes, _ := config.Dump()
			dump := string(dumpBytes)
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
