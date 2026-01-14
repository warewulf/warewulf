package set

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_Node_Set(t *testing.T) {
	tests := map[string]struct {
		args    []string
		wantErr bool
		inDB    string
		outDB   string
	}{
		"--profile=foo": {
			args:    []string{"--profile=foo", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - foo`,
		},
		"--comment=UNDEF": {
			args:    []string{"--comment=UNDEF", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: foo
    profiles:
    - default`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    profiles:
    - default`,
		},
		"--ipmiwrite=true": {
			args:    []string{"--ipmiwrite", "true", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "true"`,
		},
		"--ipmiwrite": {
			args:    []string{"--ipmiwrite", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "true"`,
		},
		"--ipmiwrite=UNDEF": {
			args:    []string{"--ipmiwrite=UNDEF", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "true"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"--ipmiwrite=false": {
			args:    []string{"--ipmiwrite=false", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    ipmi:
      write: "false"`,
		},
		"--ipmiwrite=false (override)": {
			args:    []string{"--ipmiwrite=false", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    ipmi:
      write: "true"
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default:
    ipmi:
      write: "true"
nodes:
  n01:
    profiles:
    - default
    ipmi:
      write: "false"`,
		},
		"--nettagadd": {
			args:    []string{"--nettagadd=dns=1.1.1.1", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        tags:
          dns: 1.1.1.1`,
		},
		"--nettagdel": {
			args:    []string{"--nettagdel=dns1,dns2", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n01:
    network devices:
      default:
        tags:
          dns1: 1.1.1.1
          dns2: 2.2.2.2`,
			outDB: `
nodeprofiles: {}
nodes:
  n01: {}`,
		},
		"single node change profile": {
			args:    []string{"--profile=foo", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - foo`,
		},
		"multiple nodes change profile": {
			args:    []string{"--profile=foo", "n0[1-2]"},
			wantErr: false,
			inDB: `
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
			outDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - foo
  n02:
    profiles:
    - foo`,
		},
		"single node set ipmitag": {
			args:    []string{"--ipmitagadd", "foo=baar", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    ipmi:
      tags:
        foo: baar`,
		},
		"single node delete tag": {
			args:    []string{"--tagdel", "tag1", "n01"},
			wantErr: false,
			inDB: `
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
			outDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    tags:
      tag2: value2`,
		},
		"single node add tag": {
			args:    []string{"--tagadd", "tag1=foobaar", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01: {}`,
			outDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    tags:
      tag1: foobaar`,
		},
		"single node add tag with netdev": {
			args:    []string{"--tagadd", "tag1=foobaar", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 172.16.130.101`,
			outDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    tags:
      tag1: foobaar
    network devices:
      default:
        ipaddr: 172.16.130.101`,
		},
		"single node set onboot": {
			args:    []string{"--netname", "default", "--onboot=true", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 172.16.130.101`,
			outDB: `
nodeprofiles:
  default: {}
nodes:
  n01:
    network devices:
      default:
        ipaddr: 172.16.130.101
        onboot: "true"`,
		},

		"single node set fs,part and disk": {
			args:    []string{"--fsname=var", "--fspath=/var", "--fsformat=btrfs", "--partname=var", "--partnumber=1", "--diskname=/dev/vda", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
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
        path: /var`,
		},
		"single delete not existing fs": {
			args:    []string{"--fsdel=foo", "n01"},
			wantErr: true,
			inDB: `
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
        path: /var`,
			outDB: `
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
        path: /var`,
		},
		"single node delete existing fs": {
			args:    []string{"--fsdel=/dev/disk/by-partlabel/var", "n01"},
			wantErr: false,
			inDB: `
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
        path: /var`,
			outDB: `
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
            number: "1"`,
		},
		"single node delete existing partition": {
			args:    []string{"--partdel=var", "n01"},
			wantErr: false,
			inDB: `
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
        path: /var`,
			outDB: `
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
        path: /var`,
		},
		"single node delete existing disk": {
			args:    []string{"--diskdel=/dev/vda", "n01"},
			wantErr: false,
			inDB: `
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
        path: /var`,
			outDB: `
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
        path: /var`,
		},
		"single node set mtu": {
			args:    []string{"--mtu", "1234", "--netname=mynet", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
nodeprofiles:
  default:
    comment: testit
nodes:
  n01:
    profiles:
    - default
    network devices:
      mynet:
        mtu: "1234"`,
		},
		"single node set tag": {
			args:    []string{"--tagadd", "nodetag1=nodevalue1", "n01"},
			wantErr: false,
			inDB: `
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
			outDB: `
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
      nodetag1: nodevalue1`,
		},
		"single node set comma in comment": {
			args:    []string{"n01", "--comment", "This is a , comment"},
			wantErr: false,
			inDB: `nodes:
  n01:
    comment: old comment`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    comment: This is a , comment`,
		},
		"--tagadd (one)": {
			args:    []string{"--tagadd=email=node", "n01"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    comment: testit
    tags:
      email: profile
nodes:
  n01:
    profiles:
    - default`,
			outDB: `
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
		},
		"--tagadd (second)": {
			args:    []string{"--tagadd=newtag=newval", "n01"},
			wantErr: false,
			inDB: `
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
			outDB: `
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
      newtag: newval`,
		},
		"--image=UNDEF": {
			args:    []string{"--image=UNDEF", "n1"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n1:
    image: rockylinux-9`,
			outDB: `
nodeprofiles: {}
nodes:
  n1: {}`,
		},
		"--image=UNSET": {
			args:    []string{"--image=UNSET", "n1"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n1:
    image: rockylinux-9`,
			outDB: `
nodeprofiles: {}
nodes:
  n1: {}`,
		},
		"--ipaddr=0.0.0.0 (unset)": {
			args:    []string{"--ipaddr=0.0.0.0", "n1"},
			wantErr: false,
			inDB: `
nodeprofiles: {}
nodes:
  n1:
    network devices:
      default:
        ipadddr: 192.168.0.1`,
			outDB: `
nodeprofiles: {}
nodes:
  n1: {}`,
		},
		"--partwipe": {
			args:    []string{"--partwipe", "--partname=var", "--diskname=/dev/vda", "n01"},
			wantErr: false,
			inDB: `
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"`,
			outDB: `
nodeprofiles: {}
nodes:
  n01:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
            wipe_partition_entry: true`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.WriteFile("etc/warewulf/nodes.conf", tt.inDB)
			warewulfd.SetNoDaemon()

			baseCmd := GetCommand()
			args := append(tt.args, "--yes")
			baseCmd.SetArgs(args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err := baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				content := env.ReadFile("etc/warewulf/nodes.conf")
				assert.YAMLEq(t, tt.outDB, content)
			}
		})
	}
}
