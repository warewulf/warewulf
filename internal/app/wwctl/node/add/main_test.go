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
		chkout  bool
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
`},
		{name: "single node add, profile foo",
			args:    []string{"--profile=foo", "n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - foo
`},
		{name: "single node add, discoverable true, explicit",
			args:    []string{"--discoverable=true", "n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    discoverable: "true"
    profiles:
    - default
`},
		{name: "single node add, discoverable true with yes",
			args:    []string{"--discoverable=yes", "n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    discoverable: "true"
    profiles:
    - default
`},
		{name: "single node add, discoverable wrong argument",
			args:    []string{"--discoverable=maybe", "n01"},
			wantErr: true,
			stdout:  "",
			chkout:  false,
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes: {}
`},
		{name: "single node add, discoverable false",
			args:    []string{"--discoverable=false", "n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    discoverable: "false"
    profiles:
    - default
`},
		{name: "single node add with Kernel args",
			args:    []string{"--kernelargs=foo", "n01"},
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
		{name: "single node with ipaddr6",
			args:    []string{"--ipaddr6=fdaa::1", "n01"},
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
        ip6addr: fdaa::1
`},
		{name: "single node with ipaddr",
			args:    []string{"--ipaddr=10.0.0.1", "n01"},
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
        ipaddr: 10.0.0.1
`},
		{name: "single node with malformed ipaddr",
			args:    []string{"--ipaddr=10.0.1", "n01"},
			wantErr: true,
			stdout:  "",
			chkout:  false,
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes: {}
`},
		{name: "three nodes with ipaddr",
			args:    []string{"--ipaddr=10.10.0.1", "n[01-02,03]"},
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
			args:    []string{"--ipaddr=10.10.0.1", "--netname=foo", "n[01-03]"},
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
			args:    []string{"--ipaddr=10.10.0.1", "--netname=foo", "--ipmiaddr=10.20.0.1", "n[01-03]"},
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
		{name: "one node with filesystem",
			args:    []string{"--fsname=/dev/vda1", "--fspath=/var", "n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    filesystems:
      /dev/vda1:
        path: /var
`},
		{name: "one node with filesystem",
			args:    []string{"--fsname=dev/vda1", "--fspath=/var", "n01"},
			wantErr: true,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes: {}
`},
		{name: "one node with filesystem and partition ",
			args:    []string{"--fsname=var", "--fspath=/var", "--partname=var", "--diskname=/dev/vda", "n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda:
        partitions:
          var: {}
    filesystems:
      /dev/disk/by-partlabel/var:
        path: /var
`},
		{name: "one node with filesystem with btrfs and partition ",
			args:    []string{"--fsname=var", "--fspath=/var", "--fsformat=btrfs", "--partname=var", "--diskname=/dev/vda", "n01"},
			wantErr: false,
			stdout:  "",
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda:
        partitions:
          var: {}
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`},
	}
	conf_yml := `WW_INTERNAL: 0`
	tempWarewulfConf, warewulfConfErr := os.CreateTemp("", "warewulf.conf-")
	assert.NoError(t, warewulfConfErr)
	defer os.Remove(tempWarewulfConf.Name())
	_, warewulfConfErr = tempWarewulfConf.Write([]byte(conf_yml))
	assert.NoError(t, warewulfConfErr)
	assert.NoError(t, tempWarewulfConf.Sync())
	assert.NoError(t, warewulfconf.New().Read(tempWarewulfConf.Name()))

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
			if tt.chkout && buf.String() != tt.stdout {
				t.Errorf("Got wrong output, got:'%s'\nwant:'%s'", buf.String(), tt.stdout)
				t.FailNow()
			}
		})
	}
}
