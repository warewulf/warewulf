package list

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_List(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		stdout  string
		inDb    string
	}{
		{
			name:    "single node list",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01        default
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
			name:    "multiple nodes list",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01        default
  n02        default
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
			name:    "node list returns multiple nodes",
			args:    []string{"n01,n02"},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01        default
  n02        default
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
			name:    "node list returns multiple nodes (case 2)",
			args:    []string{"n01,n03"},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01        default
  n03        default
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
			name:    "node list returns one node",
			args:    []string{"n01,"},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01        default
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
			name:    "node list profile with network",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
  n01        default         default
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
			name:    "node list profile with comment",
			args:    []string{"-a"},
			wantErr: false,
			stdout: `NODE  FIELD           PROFILE  VALUE                                         
              n01   Id              --       n01                                                  
              n01   Comment         default  profilecomment                                       
              n01   Ipxe            --       (default)                                            
              n01   RuntimeOverlay  --       (generic)                                            
              n01   SystemOverlay   --       (wwinit)                                             
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
			name:    "node list profile with comment superseded",
			args:    []string{"-a"},
			wantErr: false,
			stdout: `NODE  FIELD           PROFILE  VALUE                                         
              n01   Id              --       n01                                                  
              n01   Comment         SUPERSEDED  nodecomment                                       
              n01   Ipxe            --       (default)                                            
              n01   RuntimeOverlay  --       (generic)                                            
              n01   SystemOverlay   --       (wwinit)                                             
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
    comment: nodecomment
    profiles:
    - default
`},
		{
			name:    "node list profile with ipmi user",
			args:    []string{"-i"},
			wantErr: false,
			stdout: `NODENAME IPMIIPADDR IPMIPORT IPMIUSERNAME IPMIINTERFACE
n01 -- -- admin -- --
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
			name:    "node list profile with ipmi user superseded",
			args:    []string{"-i"},
			wantErr: false,
			stdout: `NODENAME IPMIIPADDR IPMIPORT IPMIUSERNAME IPMIINTERFACE
n01 -- -- user -- --
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
			name:    "multiple profiles list",
			args:    []string{},
			wantErr: false,
			stdout: `  NODE NAME  PROFILES  NETWORK
n01        p1,p2
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
			name:    "multiple profiles list all",
			args:    []string{"-a"},
			wantErr: false,
			stdout: `NODE  FIELD           PROFILE VALUE
n01   Id              --       n01
n01   Ipxe            --       (default)
n01   RuntimeOverlay  --       (generic)
n01   SystemOverlay   --       (wwinit) 
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
			name:    "multiple overlays list",
			args:    []string{"-l"},
			wantErr: false,
			stdout: `NODE NAME  KERNEL OVERRIDE  CONTAINER  OVERLAYS (S/R)
n01        --               --         (wwinit)/rop1,rop2
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
    - ~rop1
`,
			name:    "multiple overlays list",
			args:    []string{"-l"},
			wantErr: false,
			stdout: `NODE NAME  KERNEL OVERRIDE  CONTAINER  OVERLAYS (S/R)
n01        --               --         (wwinit)/rop2,nop1 ~{rop1}
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
    - ~rop1
`,
			name:    "multiple overlays list all",
			args:    []string{"-a"},
			wantErr: false,
			stdout: `NODE  FIELD           PROFILE     VALUE
n01   Id              --        n01        
n01   Ipxe            --          (default)
n01   RuntimeOverlay  SUPERSEDED  rop2,nop1~{rop1}
n01   SystemOverlay   --          (wwinit)  
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
		t.Logf("Running test: %s\n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			baseCmd := GetCommand()
			baseCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			old := os.Stdout // keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			err = baseCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Got unwanted error: %s", err)
				t.FailNow()
			}
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				outC <- buf.String()
			}()
			// back to normal state
			w.Close()
			os.Stdout = old // restoring the real stdout
			out := <-outC
			if strings.ReplaceAll(out, " ", "") != strings.ReplaceAll(tt.stdout, " ", "") {
				t.Errorf("Got wrong output, got:\n'%s'\nwant:\n'%s'", out, tt.stdout)
				t.FailNow()
			}
		})
	}
}
