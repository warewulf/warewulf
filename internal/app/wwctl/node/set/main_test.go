package set

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type test_description struct {
	name    string
	args    []string
	wantErr bool
	stdout  string
	inDB    string
	outDb   string
}

func run_test(t *testing.T, test test_description) {
	env := testenv.New(t)
	defer env.RemoveAll(t)
	wwlog.SetLogLevel(wwlog.DEBUG)
	env.WriteFile(t, "etc/warewulf/nodes.conf", test.inDB)
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
			content := env.ReadFile(t, "etc/warewulf/nodes.conf")
			assert.YAMLEq(t, test.outDb, content)
		}
	})
}

func Test_Single_Node_Change_Profile(t *testing.T) {
	test := test_description{
		args:    []string{"--profile=foo", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
		outDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - foo
`,
	}
	run_test(t, test)
}

func Test_Node_Unset(t *testing.T) {
	test := test_description{
		args:    []string{"--comment=UNDEF", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    comment: foo
    profiles:
    - default`,
		outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default
`,
	}
	run_test(t, test)
}

func Test_Set_Ipmi_Write_Explicit(t *testing.T) {
	test := test_description{
		args:    []string{"--ipmiwrite", "true", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01: {}
`,
		outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "true"
`,
	}
	run_test(t, test)
}

func Test_Set_Ipmi_Write_Implicit(t *testing.T) {
	test := test_description{
		args:    []string{"--ipmiwrite", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01: {}
`,
		outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "true"
`,
	}
	run_test(t, test)
}

func Test_Unset_Ipmi_Write(t *testing.T) {
	test := test_description{
		args:    []string{"--ipmiwrite=UNDEF", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "true"
`,
		outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01: {}
`,
	}
	run_test(t, test)
}

func Test_Unset_Ipmi_Write_False(t *testing.T) {
	test := test_description{
		args:    []string{"--ipmiwrite=UNDEF", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "false"
`,
		outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01: {}
`,
	}
	run_test(t, test)
}

func Test_Ipmi_Hidden_False(t *testing.T) {
	test := test_description{
		args:    []string{"--ipmiwrite=false", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    ipmi:
      write: "true"
nodes:
  n01:
    profiles:
    - default
`,
		outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    ipmi:
      write: "true"
nodes:
  n01:
    profiles:
    - default
    ipmi:
      write: "false"
`,
	}
	run_test(t, test)
}

func Test_Add_NetTags(t *testing.T) {
	test := test_description{
		args:    []string{"--nettagadd=dns=1.1.1.1", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01: {}
`,
		outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        tags:
          dns: 1.1.1.1
`,
	}
	run_test(t, test)
}

func Test_Del_NetTags(t *testing.T) {
	test := test_description{
		args:    []string{"--netname=default", "--nettagdel=dns1,dns2", "n01"},
		wantErr: false,
		stdout:  "",
		inDB: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        tags:
          dns1: 1.1.1.1
          dns2: 2.2.2.2
`,
		outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01: {}
`,
	}
	run_test(t, test)
}

func Test_Multiple_Set_Tests(t *testing.T) {
	tests := []test_description{
		{
			name:    "single node change profile",
			args:    []string{"--profile=foo", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - foo
`,
		},
		{
			name:    "multiple nodes change profile",
			args:    []string{"--profile=foo", "n0[1-2]"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
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
			outDb: `WW_INTERNAL: 45
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
`,
		},
		{
			name:    "single node set ipmitag",
			args:    []string{"--ipmitagadd", "foo=baar", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    ipmi:
      tags:
        foo: baar
`,
		},
		{
			name:    "single node delete tag",
			args:    []string{"--tagdel", "tag1", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
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
			outDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    tags:
      tag2: value2
`,
		},
		{
			name:    "single node set fs,part and disk",
			args:    []string{"--fsname=var", "--fspath=/var", "--fsformat=btrfs", "--partname=var", "--partnumber=1", "--diskname=/dev/vda", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
`,
			outDb: `WW_INTERNAL: 45
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
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
		},
		{
			name:    "single delete not existing fs",
			args:    []string{"--fsdel=foo", "n01"},
			wantErr: true,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
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
          var:
            number: "1"
        path: /var
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
			outDb: `WW_INTERNAL: 45
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
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
		},
		{
			name:    "single node delete existing fs",
			args:    []string{"--fsdel=/dev/disk/by-partlabel/var", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
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
          var:
            number: "1"
        path: /var
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
			outDb: `WW_INTERNAL: 45
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
          var:
            number: "1"
`,
		},
		{
			name:    "single node delete existing partition",
			args:    []string{"--partdel=var", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
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
          var:
            number: "1"
        path: /var
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
`,
			outDb: `WW_INTERNAL: 45
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
`,
		},
		{
			name:    "single node delete existing disk",
			args:    []string{"--diskdel=/dev/vda", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
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
			outDb: `WW_INTERNAL: 45
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
`,
		},
		{
			name:    "single node set mtu",
			args:    []string{"--mtu", "1234", "--netname=mynet", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    network devices:
      mynet:
        mtu: "1234"
`,
		},
		{
			name:    "single node set ipmitag",
			args:    []string{"--tagadd", "nodetag1=nodevalue1", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 45
nodeprofiles:
  p1:
    comment: testit 1
    tags:
      p1tag1: p1val1
  p2:
    comment: testit 1
    tags:
      p2tag2: p1val2
nodes:
  n01:
    profiles:
    - p1
    - p2`,
			outDb: `WW_INTERNAL: 45
nodeprofiles:
  p1:
    comment: testit 1
    tags:
      p1tag1: p1val1
  p2:
    comment: testit 1
    tags:
      p2tag2: p1val2
nodes:
  n01:
    profiles:
    - p1
    - p2
    tags:
      nodetag1: nodevalue1
`,
		},
		{
			name:    "single node set comma in comment",
			args:    []string{"n01", "--comment", "This is a , comment"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodes:
  n01:
    comment: old comment
`,
			outDb: `WW_INTERNAL: 43
nodeprofiles: {}
nodes:
  n01:
    comment: This is a , comment
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
	assert.NoError(t, warewulfconf.New().Read(tempWarewulfConf.Name()))

	tempNodeConf, nodesConfErr := os.CreateTemp("", "nodes.conf-")
	assert.NoError(t, nodesConfErr)
	defer os.Remove(tempNodeConf.Name())
	node.ConfigFile = tempNodeConf.Name()
	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		run_test(t, tt)
	}
}

func Test_Node_Add(t *testing.T) {
	tests := []test_description{
		{
			args:    []string{"--tagadd=email=node", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
    tags:
      email: profile
nodes:
  n01:
    profiles:
    - default`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
    tags:
      email: profile
nodes:
  n01:
    profiles:
    - default
    tags:
      email: node
`,
		},
		{
			args:    []string{"--tagadd=newtag=newval", "n01"},
			wantErr: false,
			stdout:  "",
			inDB: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
    tags:
      email: profile
nodes:
  n01:
    profiles:
    - default
    tags:
      email: node`,
			outDb: `WW_INTERNAL: 43
nodeprofiles:
  default:
    comment: testit
    tags:
      email: profile
nodes:
  n01:
    profiles:
    - default
    tags:
      email: node
      newtag: newval
`,
		},
	}

	for _, tt := range tests {
		run_test(t, tt)
	}
}
