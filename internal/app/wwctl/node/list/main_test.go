package list

import (
	"bytes"
	"os"
	"strings"

	"testing"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_List(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stdout string
		inDb   string
	}{
		{
			name: "single node list",
			args: []string{},
			stdout: `  NODE NAME  PROFILES   NETWORK  
  n01       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`,
		},
		{
			name: "multiple nodes list",
			args: []string{},
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]
  n02       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
`,
		},
		{
			name: "node list returns multiple nodes",
			args: []string{"n01,n02"},
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]           
  n02       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
`,
		},
		{
			name: "node list returns multiple nodes (case 2)",
			args: []string{"n01,n03"},
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]
  n03       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
  n03:
   profiles:
   - default
  n04:
   profiles:
   - default
  n05:
   profiles:
   - default
`,
		},
		{
			name: "node list returns one node",
			args: []string{"n01,"},
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
   profiles:
   - default
`,
		},
		{
			name: "node list profile with network",
			args: []string{},
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01       [default]        default
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    network devices:
      default:
        device: eth0
nodes:
  n01:
    profiles:
    - default
`},
		{
			name: "node list profile with comment",
			args: []string{"-a"},
			stdout: `NODE  FIELD           PROFILE  VALUE
n01   Id              --       n01
n01   Comment         default  profilecomment
n01   Ipxe            --       (default)
n01   RuntimeOverlay  --       (hosts,ssh.authorized_keys,syncuser)
n01   SystemOverlay   --       (wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,NetworkManager,debian.interfaces,wicked,ignition)
n01   Root            --       (initramfs)
n01   Init            --       (/sbin/init)
n01   Kernel.Args     --       (quiet crashkernel=no vga=791 net.naming-scheme=v238)
n01   Profiles        --       default
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: profilecomment
nodes:
  n01:
    profiles:
    - default
`},
		{
			name: "node list profile with comment superseded",
			args: []string{"-a"},
			stdout: `NODE  FIELD           PROFILE     VALUE
n01   Id              --          n01
n01   Comment         SUPERSEDED  nodecomment
n01   Ipxe            --          (default)
n01   RuntimeOverlay  --          (hosts,ssh.authorized_keys,syncuser)
n01   SystemOverlay   --          (wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,NetworkManager,debian.interfaces,wicked,ignition)
n01   Root            --          (initramfs)
n01   Init            --          (/sbin/init)
n01   Kernel.Args     --          (quiet crashkernel=no vga=791 net.naming-scheme=v238)
n01   Profiles        --          default
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    comment: profilecomment
nodes:
  n01:
    comment: nodecomment
    profiles:
    - default
`},
		{
			name: "node list profile with ipmi user",
			args: []string{"-i"},
			stdout: `NODE IPMIIPADDR IPMIPORT IPMIUSERNAME IPMIINTERFACE
n01 <nil>    admin      
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    ipmi:
      username: admin
nodes:
  n01:
    profiles:
    - default
`},
		{
			name: "node list profile with ipmi user superseded",
			args: []string{"-i"},
			stdout: `NODE IPMIIPADDR IPMIPORT IPMIUSERNAME IPMIINTERFACE
n01 <nil>    user
`,
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  default:
    ipmi:
      username: admin
nodes:
  n01:
    ipmi:
      username: user
    profiles:
    - default
`},
		{
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  p1: {}
  p2: {}
nodes:
  n01:
    profiles:
    - p1
    - p2
`,
			name: "multiple profiles list",
			args: []string{},
			stdout: `  NODE NAME  PROFILES  NETWORK
n01        [p1 p2]
`},
		{
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  p1: {}
  p2: {}
nodes:
  n01:
    profiles:
    - p1
    - p2
`,
			name: "multiple profiles list all",
			args: []string{"-a"},
			stdout: `NODE  FIELD           PROFILE  VALUE
n01   Id              --       n01
n01   Ipxe            --       (default)
n01   RuntimeOverlay  --       (hosts,ssh.authorized_keys,syncuser)
n01   SystemOverlay   --       (wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,NetworkManager,debian.interfaces,wicked,ignition)
n01   Root            --       (initramfs)
n01   Init            --       (/sbin/init)
n01   Kernel.Args     --       (quiet crashkernel=no vga=791 net.naming-scheme=v238)
n01   Profiles        --       p1,p2
`},
		{
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  p1:
    runtime overlay:
    - rop1
    - rop2
nodes:
  n01:
    profiles:
    - p1
`,
			name: "multiple overlays list",
			args: []string{"-l"},
			stdout: `NODE NAME  KERNEL OVERRIDE  CONTAINER  OVERLAYS (S/R)
n01        --               --         (wwinit,wwclient,fstab,hostname,ssh.host_keys,issue,resolv,udev.netname,systemd.netname,ifcfg,NetworkManager,debian.interfaces,wicked,ignition)/rop1,rop2
`},
		{
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  p1:
    system overlay:
    - sop1
    runtime overlay:
    - rop1
    - rop2
nodes:
  n01:
    profiles:
    - p1
    runtime overlay:
    - nop1
    - ~rop1
`,
			name: "multiple overlays list",
			args: []string{"-l"},
			stdout: `NODE NAME  KERNEL OVERRIDE  CONTAINER  OVERLAYS (S/R)
n01        --               --         sop1/rop2,nop1 ~{rop1}
`},
		{
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  p1:
    system overlay:
    - sop1
    runtime overlay:
    - rop1
    - rop2
nodes:
  n01:
    profiles:
    - p1
    runtime overlay:
    - nop1
    - ~rop1
`,
			name: "multiple overlays list all",
			args: []string{"-a"},
			stdout: `NODE  FIELD           PROFILE     VALUE
n01   Id              --          n01
n01   Ipxe            --          (default)
n01   RuntimeOverlay  SUPERSEDED  rop2,nop1 ~{rop1}
n01   SystemOverlay   p1          sop1
n01   Root            --          (initramfs)
n01   Init            --          (/sbin/init)
n01   Kernel.Args     --          (quiet crashkernel=no vga=791 net.naming-scheme=v238)
n01   Profiles        --          p1
`},
		{
			inDb: `WW_INTERNAL: 45
nodeprofiles:
  p1:
    runtime overlay:
    - rop1
    - rop2
nodes:
  n01:
    profiles:
    - p1
    runtime overlay:
    - nop1
`,
			name: "multiple overlays list all",
			args: []string{"-a"},
			stdout: `NODE  FIELD           PROFILE     VALUE
n01   Id              --          n01
n01   Ipxe            --          (default)
n01   RuntimeOverlay  --          (hosts,ssh.authorized_keys,syncuser)
n01   SystemOverlay   SUPERSEDED  profileinit,nodeinit
n01   Root            --          (initramfs)
n01   Init            --          (/sbin/init)
n01   Kernel.Args     --          (quiet crashkernel=no vga=791 net.naming-scheme=v238)
n01   Profiles        --          p1
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
		_, err = tempNodeConf.Write([]byte(tt.inDb))
		assert.NoError(t, err)
		assert.NoError(t, tempNodeConf.Sync())

		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			//verifyOutput(t, baseCmd, tt.stdout)
			buf := new(bytes.Buffer)
			wwlog.SetLogWriter(buf)
			wwlog.SetLogWriterErr(buf)
			wwlog.SetLogWriterInfo(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			assert.NotEmpty(t, buf, "output should not be empty")
			assert.Contains(t, strings.ReplaceAll(buf.String(), " ", ""),
				strings.ReplaceAll(tt.stdout, " ", ""))

		})
	}
}

func TestListMultipleFormats(t *testing.T) {
	t.Skip("temporally skip this test")
	tests := []struct {
		name    string
		args    []string
		stdout  string
		inDb    string
		wantErr bool
	}{
		{
			name:   "single node list yaml output",
			args:   []string{"-y"},
			stdout: "n01:\n  AssetKey: \"\"\n  ClusterName: \"\"\n  Comment: \"\"\n  ContainerName: \"\"\n  Discoverable: \"\"\n  Disks: {}\n  FileSystems: {}\n  Grub: \"\"\n  Id: |\n    Source: explicit\n    Value: n01\n  Init: |\n    Source: default-value\n    Value: /sbin/init\n  Ipmi:\n    EscapeChar: \"\"\n    Gateway: \"\"\n    Interface: \"\"\n    Ipaddr: \"\"\n    Netmask: \"\"\n    Password: \"\"\n    Port: \"\"\n    Tags: null\n    UserName: \"\"\n    Write: \"\"\n  Ipxe: |\n    Source: default-value\n    Value: default\n  Kernel:\n    Args: |\n      Source: default-value\n      Value: quiet crashkernel=no vga=791 net.naming-scheme=v238\n    Override: \"\"\n  NetDevs: {}\n  PrimaryNetDev: \"\"\n  Profiles: |\n    Source: explicit\n    Value: default\n  Root: |\n    Source: default-value\n    Value: initramfs\n  RuntimeOverlay: |\n    Source: default-value\n    Value: generic\n  SystemOverlay: |\n    Source: default-value\n    Value: wwinit\n  Tags: {}\n",
			inDb: `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`,
		},
		{
			name:   "single node list json output",
			args:   []string{"-j"},
			stdout: "{\"n01\":{\"Id\":\"Source: explicit\\nValue: n01\\n\",\"Comment\":\"\",\"ClusterName\":\"\",\"ContainerName\":\"\",\"Ipxe\":\"Source: default-value\\nValue: default\\n\",\"Grub\":\"\",\"RuntimeOverlay\":\"Source: default-value\\nValue: generic\\n\",\"SystemOverlay\":\"Source: default-value\\nValue: wwinit\\n\",\"Root\":\"Source: default-value\\nValue: initramfs\\n\",\"Discoverable\":\"\",\"Init\":\"Source: default-value\\nValue: /sbin/init\\n\",\"AssetKey\":\"\",\"Kernel\":{\"Override\":\"\",\"Args\":\"Source: default-value\\nValue: quiet crashkernel=no vga=791 net.naming-scheme=v238\\n\"},\"Ipmi\":{\"Ipaddr\":\"\",\"Netmask\":\"\",\"Port\":\"\",\"Gateway\":\"\",\"UserName\":\"\",\"Password\":\"\",\"Interface\":\"\",\"EscapeChar\":\"\",\"Write\":\"\",\"Tags\":null},\"Profiles\":\"Source: explicit\\nValue: default\\n\",\"PrimaryNetDev\":\"\",\"NetDevs\":{},\"Tags\":{},\"Disks\":{},\"FileSystems\":{}}}\n",
			inDb: `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
`,
		},
		{
			name:   "multiple nodes list json output",
			args:   []string{"-j"},
			stdout: "n01  n02",
			inDb: `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
    profiles:
    - default
`,
		},
		{
			name:   "multiple nodes list yaml output",
			args:   []string{"-y"},
			stdout: "n01: n02:",
			inDb: `WW_INTERNAL: 43
nodeprofiles:
  default: {}
nodes:
  n01:
    profiles:
    - default
  n02:
    profiles:
    - default
`,
		},
	}

	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		env := testenv.New(t)
		env.WriteFile(t, "etc/warewulf/nodes.conf", tt.inDb)
		var err error
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			wwlog.SetLogWriter(buf)
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			wwlog.SetLogWriter(buf)
			err := baseCmd.Execute()
			assert.NoError(t, err)
			for _, expected_output := range tt.output {
				assert.Equal(t, expected_output, buf.String())
			}
		})
	}
}

func verifyOutput(t *testing.T, baseCmd *cobra.Command, content string) {
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	wwlog.SetLogWriter(os.Stdout)
	baseCmd.SetOut(os.Stdout)
	baseCmd.SetErr(os.Stdout)
	err := baseCmd.Execute()
	assert.NoError(t, err)

	strip, _ := regexp.Compile("(?m)(^  *|  *$)")

	stdoutC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR)
		stdoutC <- buf.String()
	}()
	stdoutW.Close()

	stdout := <-stdoutC
	assert.NotEmpty(t, stdout, "output should not be empty")
	assert.Equal(t, strip.ReplaceAllString(content, ""), strip.ReplaceAllString(stdout, ""))
}
