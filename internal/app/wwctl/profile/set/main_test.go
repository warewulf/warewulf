package set

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
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
