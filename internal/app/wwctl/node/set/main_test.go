package set

import (
	"bytes"
	"os"
	"testing"

	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/ww4test"
	"github.com/stretchr/testify/assert"
)

type test_description struct {
	name    string
	args    []string
	wantErr bool
	stdout  string
	outDb   string
	inDB    string
}

func run_test(t *testing.T, test test_description) {
	//wwlog.SetLogLevel(wwlog.DEBUG)
	var env ww4test.WarewulfTestEnv
	env.NodesConf = test.inDB
	env.New(t)
	defer os.RemoveAll(env.BaseDir)
	warewulfd.SetNoDaemon()
	name := test.name
	if name == "" {
		name = t.Name()
	}
	t.Run(name, func(t *testing.T) {
		baseCmd := GetCommand()
		test.args = append(test.args, "--yes")
		baseCmd.SetArgs(test.args)
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		err := baseCmd.Execute()
		if test.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, buf.String(), test.stdout)
			content, err := os.ReadFile(env.NodesConfFile)
			assert.NoError(t, err)
			assert.Equal(t, test.outDb, string(content))
		}
	})
}

func Test_Single_Node_Change_Profile(t *testing.T) {
	test := test_description{
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
`}
	run_test(t, test)
}

func Test_Multiple_Add_Tests(t *testing.T) {
	tests := []test_description{
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
	for _, tt := range tests {
		run_test(t, tt)
	}
}
