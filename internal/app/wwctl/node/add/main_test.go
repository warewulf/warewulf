package add

import (
	"bytes"
	"fmt"
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
		flags   map[string]string
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
`},
		{name: "single node add, profile foo",
			args:    []string{"n01"},
			flags:   map[string]string{"profile": "foo"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - foo
`},
		{name: "single node add with Kernel args",
			args:    []string{"n01"},
			flags:   map[string]string{"kernelargs": "foo"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    kernel:
      args: foo
    profiles:
    - default
`},
		{name: "double node add explicit",
			args:    []string{"n01", "n02"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
  n02:
    profiles:
    - default
`},
		{name: "single node with ipaddr",
			args:    []string{"n01"},
			flags:   map[string]string{"ipaddr": "10.10.0.1"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.10.0.1
`},
		{name: "three nodes with ipaddr",
			args:    []string{"n[01-02,03]"},
			flags:   map[string]string{"ipaddr": "10.10.0.1"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.10.0.1
  n02:
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.10.0.2
  n03:
    profiles:
    - default
    network devices:
      default:
        ipaddr: 10.10.0.3
`},
		{name: "three nodes with ipaddr different network",
			args:    []string{"n[01-03]"},
			flags:   map[string]string{"ipaddr": "10.10.0.1", "netname": "foo"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    network devices:
      foo:
        ipaddr: 10.10.0.1
  n02:
    profiles:
    - default
    network devices:
      foo:
        ipaddr: 10.10.0.2
  n03:
    profiles:
    - default
    network devices:
      foo:
        ipaddr: 10.10.0.3
`},
		{name: "three nodes with ipaddr different network, with ipmiaddr",
			args:    []string{"n[01-03]"},
			flags:   map[string]string{"ipaddr": "10.10.0.1", "netname": "foo", "ipmiaddr": "10.20.0.1"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      ipaddr: 10.20.0.1
    profiles:
    - default
    network devices:
      foo:
        ipaddr: 10.10.0.1
  n02:
    ipmi:
      ipaddr: 10.20.0.2
    profiles:
    - default
    network devices:
      foo:
        ipaddr: 10.10.0.2
  n03:
    ipmi:
      ipaddr: 10.20.0.3
    profiles:
    - default
    network devices:
      foo:
        ipaddr: 10.10.0.3
`},
	}
	for _, tt := range tests {
		db, err = node.TestNew([]byte(nodes_yml))
		assert.NoError(t, err)
		fmt.Printf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			// store global NodeConf as the NodeConf.NetDevs["default"] will be delete
			// in the main.go
			tmpConfNet := NodeConf.NetDevs["default"]
			//tmpConf := NodeConf
			//tmpKernel := NodeConf.Kernel
			baseCmd.SetArgs(tt.args)
			for key, val := range tt.flags {
				baseCmd.Flags().Set(key, val)
			}
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
				return
			}
			NodeConf.NetDevs["default"] = tmpConfNet
			for key, _ := range tt.flags {
				baseCmd.Flags().Set(key, "hark")
			}

			//NodeConf = tmpConf
			//NodeConf.Kernel = node.NewConf().Kernel
		})
	}
}
