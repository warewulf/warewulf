package set

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/profile/list"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_Profile_Set(t *testing.T) {
	tests := map[string]struct {
		args    []string
		wantErr bool
		inDB    string
		outDb   string
	}{
		"Test_Set_Netdev": {
			args:    []string{"--netdev=eth0", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes: {}`,
		},
		"Test_Set_Netdev_and_Mask": {
			args:    []string{"--netdev=eth0", "-M=255.255.255.0", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        netmask: 255.255.255.0
nodes: {}`,
		},
		"Set Mask Existing NetDev": {
			args:    []string{"-M=255.255.255.0", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        netmask: 255.255.255.0
nodes: {}`,
		},
		"--image=UNSET": {
			args:    []string{"--image=UNSET", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    image: rockylinux-9
nodes: {}`,
			outDb: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"--image=UNDEF": {
			args:    []string{"--image=UNDEF", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    image: rockylinux-9
nodes: {}`,
			outDb: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"--tagadd=mytag=0.0.0.0": {
			args:    []string{"--tagadd=mytag=0.0.0.0", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    tags:
      mytag: 0.0.0.0
nodes: {}`,
		},
		"set fs,part and disk": {
			args: []string{"--fsname=var", "--fspath=/var", "--fsformat=btrfs", "--partname=var", "--partnumber=1", "--diskname=/dev/vda", "default"},
			inDB: `
nodeprofiles:
  default: {}
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}`,
		},
		"single delete not existing fs": {
			args:    []string{"--fsdel=foo", "default"},
			wantErr: true,
			inDB: `
nodeprofiles:
  default: {}
nodes: {} `,
			outDb: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"single node delete existing partition": {
			args:    []string{"--partdel=var", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}
`,
			outDb: `
nodeprofiles:
  default:
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}`,
		},
		"single profile delete partition scoped to diskname": {
			args:    []string{"--partdel=var", "--diskname=/dev/vda", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
      /dev/vdb:
        partitions:
          var:
            number: "1"
nodes: {}
`,
			outDb: `
nodeprofiles:
  default:
    disks:
      /dev/vda: {}
      /dev/vdb:
        partitions:
          var:
            number: "1"
nodes: {}`,
		},
		"single profile delete partition unscoped removes from all disks": {
			args:    []string{"--partdel=var", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
      /dev/vdb:
        partitions:
          var:
            number: "1"
nodes: {}
`,
			outDb: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"single profile delete partition unscoped skips disks without it": {
			args:    []string{"--partdel=var", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
      /dev/vdb:
        partitions:
          boot:
            number: "2"
nodes: {}
`,
			outDb: `
nodeprofiles:
  default:
    disks:
      /dev/vda: {}
      /dev/vdb:
        partitions:
          boot:
            number: "2"
nodes: {}`,
		},
		"single profile delete partition unscoped errors when not found": {
			args:    []string{"--partdel=nonexistent", "default"},
			wantErr: true,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
nodes: {}
`,
		},
		"single set wipetabe to true": {
			args:    []string{"--diskwipe=true", "--partname=var", "--diskname=/dev/vda", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}
`,
			outDb: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        wipe_table: true
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}`,
		},
		"single set wipetabe to false": {
			args:    []string{"--diskwipe=false", "--partname=var", "--diskname=/dev/vda", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        wipe_table: true
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}
`,
			outDb: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        wipe_table: false
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}`,
		},
		"single set partwipe to true": {
			args:    []string{"--partwipe=true", "--partname=var", "--diskname=/dev/vda", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}
`,
			outDb: `
nodeprofiles:
  default:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
            wipe_partition_entry: true
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
nodes: {}`,
		},
		"--nettagadd with existing netdev": {
			args:    []string{"--netname=eth0", "--nettagadd=DNS1=1.1.1.1", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      eth0:
        device: eth0
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      eth0:
        device: eth0
        tags:
          DNS1: "1.1.1.1"
nodes: {}`,
		},
		"--nettagadd adds tags to default profile": {
			args:    []string{"--nettagadd=DNS1=1.1.1.1", "--nettagadd=DNS2=1.0.0.1", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        tags:
          DNS1: "1.1.1.1"
          DNS2: "1.0.0.1"
nodes: {}`,
		},
		"--tagadd does not add network tags": {
			args:    []string{"--tagadd=foo=bar", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
    tags:
      foo: bar
nodes: {}`,
		},
		"--tagadd with --netdev does not add network tags": {
			args:    []string{"--tagadd=foo=bar", "--netdev=eth1", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth1
        tags:
          DNS1: 1.1.1.1
    tags:
      foo: bar
nodes: {}`,
		},
		"--tagadd with --nettagadd": {
			args:    []string{"--tagadd=foo=bar", "--nettagadd=DNS2=1.0.0.1", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
          DNS2: 1.0.0.1
    tags:
      foo: bar
nodes: {}`,
		},
		"--tagadd with new --netname": {
			args:    []string{"--tagadd=foo=bar", "--netname=ib0", "--netdev=ib0", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
      ib0:
        device: ib0
    tags:
      foo: bar
nodes: {}`,
		},
		"--nettagdel does not delete tags": {
			args:    []string{"--nettagdel=DNS1", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
    tags:
      DNS1: 1.1.1.1
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
    tags:
      DNS1: 1.1.1.1
nodes: {}`,
		},
		"--netdel with --nettagadd": {
			args:    []string{"--netname=ib0", "--netdel=ib0", "--nettagadd=mtu=9000", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      ib0:
        device: ib0
nodes: {}`,
			outDb: `
nodeprofiles:
  default: {}
nodes: {}`,
		},
		"deprecated aliases": {
			args:    []string{"--container=img", "--wwinit=sys", "--runtime=rt", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default: {}
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    image name: img
    system overlay:
      - sys
    runtime overlay:
      - rt
nodes: {}`,
		},
		"--tagdel does not delete network tags": {
			args:    []string{"--tagdel=DNS1", "default"},
			wantErr: false,
			inDB: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
    tags:
      DNS1: 1.1.1.1
nodes: {}`,
			outDb: `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
        tags:
          DNS1: 1.1.1.1
nodes: {}`,
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
				assert.YAMLEq(t, tt.outDb, content)
			}
		})
	}
}

// Test_Profile_Set_Tagadd_List checks that a tag set with --tagadd is listed
// only as a profile tag, and not also as a network tag. (#2292)
func Test_Profile_Set_Tagadd_List(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes: {}`)
	warewulfd.SetNoDaemon()

	setCmd := GetCommand()
	setCmd.SetArgs([]string{"--tagadd=foo=bar", "default", "--yes"})
	setCmd.SetOut(new(bytes.Buffer))
	setCmd.SetErr(new(bytes.Buffer))
	assert.NoError(t, setCmd.Execute())

	buf := new(bytes.Buffer)
	wwlog.SetLogWriter(buf)
	defer wwlog.SetLogWriterInfo(os.Stdout)
	defer wwlog.SetLogWriterErr(os.Stderr)
	listCmd := list.GetCommand()
	listCmd.SetArgs([]string{"-a", "default"})
	listCmd.SetOut(buf)
	listCmd.SetErr(buf)
	assert.NoError(t, listCmd.Execute())
	assert.Regexp(t, regexp.MustCompile(`(?m)^default\s+Tags\[foo\]\s+bar\s*$`), buf.String())
	assert.NotContains(t, buf.String(), "NetDevs[default].Tags[foo]")
}

func Test_Profile_Set_Nettagdel_Warning(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes: {}`)
	warewulfd.SetNoDaemon()
	buf := new(bytes.Buffer)
	wwlog.SetLogWriter(buf)
	defer wwlog.SetLogWriterInfo(os.Stdout)
	defer wwlog.SetLogWriterErr(os.Stderr)
	baseCmd := GetCommand()
	baseCmd.SetArgs([]string{"--netname=ib0", "--nettagdel=DNS1", "default", "--yes"})
	baseCmd.SetOut(buf)
	baseCmd.SetErr(buf)
	assert.NoError(t, baseCmd.Execute())
	assert.Contains(t, buf.String(), "Profile: default, network device ib0 is not defined on this profile; no network tags deleted")
}
