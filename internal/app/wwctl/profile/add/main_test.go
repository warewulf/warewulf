package add

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
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
			args:    []string{"p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01: {}
nodes: {}
`,
		},
		{
			name:    "single profile add with netname and netdev",
			args:    []string{"--netname", "primary", "--netdev", "eno3", "p02"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p02:
    network devices:
      primary:
        device: eno3
nodes: {}
`,
		},
		{
			name:    "single profile add with tags",
			args:    []string{"--tagadd=foo=bar", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    tags:
      foo: bar
nodes: {}
`,
		},
		{
			name:    "single profile add with network tags",
			args:    []string{"--nettagadd=dns=1.1.1.1", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    network devices:
      default:
        tags:
          dns: 1.1.1.1
nodes: {}
`,
		},
		{
			name:    "single profile add with netdev, tags, and network tags",
			args:    []string{"--netdev=eth0", "--tagadd=foo=bar", "--nettagadd=dns=1.1.1.1", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    network devices:
      default:
        device: eth0
        tags:
          dns: 1.1.1.1
    tags:
      foo: bar
nodes: {}
`,
		},
		{
			name:    "single profile add with ipmi tags",
			args:    []string{"--ipmitagadd=foo=bar", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    ipmi:
      tags:
        foo: bar
nodes: {}
`,
		},
		{
			name:    "single profile add with deprecated aliases",
			args:    []string{"--container=img", "--wwinit=sys", "--runtime=rt", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    image name: img
    system overlay:
      - sys
    runtime overlay:
      - rt
nodes: {}
`,
		},
		{
			name:    "multiple profile add with comment and cluster",
			args:    []string{"--comment=hello", "--cluster=c1", "p01", "p02"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    comment: hello
    cluster name: c1
  p02:
    comment: hello
    cluster name: c1
nodes: {}
`,
		},
		{
			name:    "single profile add with image and overlays",
			args:    []string{"--image=img", "--system-overlays=a,b", "--runtime-overlays=c", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    image name: img
    system overlay:
      - a
      - b
    runtime overlay:
      - c
nodes: {}
`,
		},
		{
			name:    "single profile add with kernel",
			args:    []string{"--kernelversion=6.1", "--kernelargs=quiet,crashkernel=no", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    kernel:
      version: "6.1"
      args:
        - quiet
        - crashkernel=no
nodes: {}
`,
		},
		{
			name:    "single profile add with ipmi",
			args:    []string{"--ipmiuser=u", "--ipmipass=p", "--ipmiaddr=10.0.0.5", "--ipmigateway=10.0.0.1", "--ipminetmask=255.255.255.0", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    ipmi:
      username: u
      password: p
      ipaddr: 10.0.0.5
      gateway: 10.0.0.1
      netmask: 255.255.255.0
nodes: {}
`,
		},
		{
			name:    "single profile add with network",
			args:    []string{"--netname=primary", "--netdev=eno3", "--hwaddr=aa:bb:cc:dd:ee:ff", "--ipaddr=10.0.0.10", "--netmask=255.255.255.0", "--gateway=10.0.0.1", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    network devices:
      primary:
        device: eno3
        hwaddr: aa:bb:cc:dd:ee:ff
        ipaddr: 10.0.0.10
        netmask: 255.255.255.0
        gateway: 10.0.0.1
nodes: {}
`,
		},
		{
			name:    "single profile add with partition",
			args:    []string{"--diskname=/dev/vda", "--partname=var", "--parttype=8300", "--partwipe", "--partcreate=true", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    disks:
      /dev/vda:
        partitions:
          var:
            type_guid: "8300"
            wipe_partition_entry: true
            should_exist: true
nodes: {}
`,
		},
		{
			name:    "single profile add with filesystem on partition",
			args:    []string{"--diskname=/dev/vda", "--partname=var", "--partnumber=1", "--fsname=var", "--fspath=/var", "--fsformat=btrfs", "--fswipe=true", "p01"},
			wantErr: false,
			stdout:  "",
			outDb: `nodeprofiles:
  p01:
    disks:
      /dev/vda:
        partitions:
          var:
            number: "1"
    filesystems:
      /dev/disk/by-partlabel/var:
        format: btrfs
        path: /var
        wipe_filesystem: true
nodes: {}
`,
		},
	}

	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		env := testenv.New(t)
		env.WriteFile("etc/warewulf/nodes.conf", ``)
		var err error
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err = baseCmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			config, configErr := node.New()
			assert.NoError(t, configErr)
			dumpBytes, _ := config.Dump()
			assert.YAMLEq(t, tt.outDb, string(dumpBytes))
		})
	}
}
