package set

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
		inDB    string
	}{
		{name: "single node change profile",
			args:    []string{"--profile=foo", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - foo
`},
		{name: "multiple nodes change profile",
			args:    []string{"--profile=foo", "n0[1-2]"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
  n02:
    profiles:
    - default`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - foo
  n02:
    profiles:
    - foo
`},
		{name: "single node set ipmitag",
			args:    []string{"--ipmitagadd", "foo=baar", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    ipmi:
      tags:
        foo: baar
    profiles:
    - default
`},
		{name: "single node delete tag",
			args:    []string{"--tagdel", "tag1", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    tags:
      tag1: value1
      tag2: value2`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    tags:
      tag2: value2
`},
		{name: "single node set fs,part and disk",
			args:    []string{"--fsname=var", "--fspath=/var", "--fsformat=btrfs", "--partname=var", "--diskname=/dev/vda", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
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
		{name: "single delete not existing fs",
			args:    []string{"--fsdel=var", "n01"},
			wantErr: true,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda:
        partitions:
          var: {}
        path: /var
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
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
		{name: "single node delete existing fs",
			args:    []string{"--fsdel=/dev/disk/by-partlabel/var", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda:
        partitions:
          var: {}
        path: /var
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda:
        partitions:
          var: {}
`},
		{name: "single node delete existing partition",
			args:    []string{"--partdel=var", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda:
        partitions:
          var: {}
        path: /var
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda: {}
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`},
		{name: "single node delete existing disk",
			args:    []string{"--diskdel=/dev/vda", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    disks:
      /dev/vda:
        partitions:
          var: {}
        path: /var
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
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
		_, err = tempNodeConf.Write([]byte(tt.inDB))
		assert.NoError(t, err)
		assert.NoError(t, tempNodeConf.Sync())
		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			tt.args = append(tt.args, "--yes")
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
