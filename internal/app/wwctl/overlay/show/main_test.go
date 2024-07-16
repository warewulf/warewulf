package show

import (
	"bytes"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

var (
	overlayEmail = `
{{ if .Tags.email }}eMail: {{ .Tags.email }}{{else}} noMail{{- end }}
`
	overlayOverlay = `
overlay name {{ .Overlay }}
`
)

func Test_Overlay_List(t *testing.T) {
	env := testenv.New(t)
	env.WriteFile(t, "etc/warewulf/nodes.conf",
		`WW_INTERNAL: 45
nodeprofiles:
  default:
    tags:
      email: admin@localhost
  empty: {}
nodes:
  node1:
    tags:
      email: admin@node1
  node2: {}
  node3:
    profiles:
      - empty
`)

	env.WriteFile(t, path.Join(testenv.WWOverlaydir, "testoverlay/email.ww"), overlayEmail)
	env.WriteFile(t, path.Join(testenv.WWOverlaydir, "testoverlay/overlay.ww"), overlayOverlay)
	defer env.RemoveAll(t)
	warewulfd.SetNoDaemon()
	t.Run("overlay show raw", func(t *testing.T) {
		baseCmd.SetArgs([]string{"testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), overlayEmail)
	})
	t.Run("overlay show rendered node tag", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "node1", "testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "admin@node1")
	})
	t.Run("overlay show rendered profile tag", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "node2", "testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "admin@localhost")
	})
	t.Run("overlay show no tag", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "node3", "testoverlay", "email.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "noMail")
	})
	t.Run("overlay shows overlay", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "node1", "testoverlay", "overlay.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "testoverlay")
	})
}

func TestShowServerTemplate(t *testing.T) {
	const template = `
	Id: {{.Id}}
	ClusterName: {{.ClusterName}}
	BuildHost: {{.BuildHost}}
	`

	env := testenv.New(t)
	env.WriteFile(t, "etc/warewulf/nodes.conf",
		`WW_INTERNAL: 45
nodeprofiles:
  default:
    tags:
      email: admin@localhost
  empty: {}
nodes:
  node1:
    tags:
      email: admin@node1
  node2: {}
  node3:
    profiles:
      - empty
`)

	env.WriteFile(t, path.Join(testenv.WWOverlaydir, "testoverlay/template.ww"), template)
	defer env.RemoveAll(t)
	warewulfd.SetNoDaemon()

	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("overlay render host template using 'host' value", func(t *testing.T) {
		baseCmd.SetArgs([]string{"-r", "host", "testoverlay", "template.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err := baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Id: "+host)
		assert.Contains(t, buf.String(), "ClusterName: "+host)
		assert.Contains(t, buf.String(), "BuildHost: "+host)
	})

	t.Run("overlay render host template using host domain value", func(t *testing.T) {
		host, err := os.Hostname()
		if err != nil {
			t.Fatal(err)
		}
		baseCmd.SetArgs([]string{"-r", host, "testoverlay", "template.ww"})
		baseCmd := GetCommand()
		buf := new(bytes.Buffer)
		baseCmd.SetOut(buf)
		baseCmd.SetErr(buf)
		wwlog.SetLogWriter(buf)
		err = baseCmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Id: "+host)
		assert.Contains(t, buf.String(), "ClusterName: "+host)
		assert.Contains(t, buf.String(), "BuildHost: "+host)
	})
}

func Test_RenderTemplates(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll(t)
	env.WriteFile(t, "etc/warewulf/warewulf.conf",
		`WW_INTERNAL: 43
ipaddr: 192.168.0.1/24
netmask: 255.255.255.0
network: 192.168.0.0
warewulf:
  port: 9873
  secure: false
  update interval: 60
  autobuild overlays: true
  host overlay: true
  syslog: true
dhcp:
  enabled: false
tftp:
  enabled: false
nfs:
  enabled: false`)
	env.WriteFile(t, "etc/warewulf/nodes.conf",
		`WW_INTERNAL: 45
nodeprofiles:
  empty: {}
nodes:
  node1:
    profiles:
    - empty
    network devices:
      default:
        device: wwnet0
        hwaddr: e6:92:39:49:7b:03
        ipaddr: 192.168.3.21
        netmask: 255.255.255.0
        gateway: 192.168.3.1
      secondary:
        device: wwnet1
        hwaddr: 9a:77:29:73:14:f1
        ipaddr: 192.168.3.22
        netmask: 255.255.255.0
        gateway: 192.168.3.1
        tags:
          DNS1: 8.8.8.8
          DNS2: 8.8.4.4
    ipmi:
      username: user
      password: password
      ipaddr: 192.168.4.21
      netmask: 255.255.255.0
      gateway: 192.168.4.1
      write: "true"
`)
	env.ImportFile(t, "var/lib/warewulf/overlays/ifcfg/rootfs/etc/sysconfig/network-scripts/ifcfg.ww", "../../../../../overlays/ifcfg/rootfs/etc/sysconfig/network-scripts/ifcfg.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ifcfg/rootfs/etc/sysconfig/network.ww", "../../../../../overlays/ifcfg/rootfs/etc/sysconfig/network.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/wwinit/rootfs/etc/warewulf/warewulf.conf.ww", "../../../../../overlays/wwinit/rootfs/etc/warewulf/warewulf.conf.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/wwinit/rootfs/warewulf/config.ww", "../../../../../overlays/wwinit/rootfs/warewulf/config.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/wwinit/rootfs/warewulf/init.d/80-wwclient.ww", "../../../../../overlays/wwinit/rootfs/warewulf/init.d/80-wwclient.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/ssh_authorized_keys/rootfs/root/.ssh/authorized_keys.ww", "../../../../../overlays/ssh_authorized_keys/rootfs/root/.ssh/authorized_keys.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "ifcfg:ifcfg.ww",
			args: []string{"--render", "node1", "ifcfg", "etc/sysconfig/network-scripts/ifcfg.ww"},
			log:  ifcfg,
		},
		{
			name: "ifcfg:network.ww",
			args: []string{"--render", "node1", "ifcfg", "etc/sysconfig/network.ww"},
			log:  ifcfg_network,
		},
		{
			name: "wwinit:warewulf.conf.ww",
			args: []string{"--render", "node1", "wwinit", "etc/warewulf/warewulf.conf.ww"},
			log:  wwinit_warewulf_conf,
		},
		{
			name: "wwinit:config.ww",
			args: []string{"--render", "node1", "wwinit", "warewulf/config.ww"},
			log:  wwinit_config,
		},
		{
			name: "wwinit:wwclient.ww",
			args: []string{"--render", "node1", "wwinit", "warewulf/init.d/80-wwclient.ww"},
			log:  wwinit_wwclient,
		},
		// {
		// 	name: "ssh_authorized_keys:authorized_keys.ww",
		// 	args: []string{"--render", "node1", "ssh_authorized_keys", "root/.ssh/authorized_keys.ww"},
		// 	log:  ssh_authorized_keys,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetCommand()
			cmd.SetArgs(tt.args)
			stdout := bytes.NewBufferString("")
			stderr := bytes.NewBufferString("")
			logbuf := bytes.NewBufferString("")
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			wwlog.SetLogWriter(logbuf)
			err := cmd.Execute()
			assert.NoError(t, err)
			assert.Empty(t, stdout.String())
			assert.Empty(t, stderr.String())
			assert.Equal(t, tt.log, cleanHeader(logbuf.String()))
		})
	}
}

func cleanHeader(input string) (output string) {
	host := regexp.MustCompile(`# Host:.*`)
	time := regexp.MustCompile(`# Time:.*`)
	source := regexp.MustCompile(`/ww4test-[0-9]*/`)

	output = input
	output = host.ReplaceAllString(output, "# Host:   REMOVED")
	output = time.ReplaceAllString(output, "# Time:   REMOVED")
	output = source.ReplaceAllString(output, "/ww4test-REMOVED/")
	return output
}

// overlays/syncuser/rootfs/etc/group.ww
// overlays/syncuser/rootfs/etc/passwd.ww

// overlays/hosts/rootfs/etc/hosts.ww

// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_dsa_key.pub.ww
// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_dsa_key.ww
// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.pub.ww
// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.ww
// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.pub.ww
// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.ww
// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_rsa_key.pub.ww
// overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_rsa_key.ww

// overlays/NetworkManager/rootfs/etc/NetworkManager/conf.d/ww4-unmanaged.ww
// overlays/NetworkManager/rootfs/etc/NetworkManager/system-connections/ww4-managed.ww

// overlays/fstab/rootfs/etc/fstab.ww

// overlays/hostname/rootfs/etc/hostname.ww

// overlays/issue/rootfs/etc/issue.ww

// overlays/resolv/rootfs/etc/resolv.conf.ww

// overlays/wicked/rootfs/etc/wicked/ifconfig/ifcfg.xml.ww

// overlays/debian.interfaces/rootfs/etc/network/interfaces.d/default.ww

// overlays/udev/rootfs/etc/udev/rules.d/70-ww4-netname.rules.ww

// overlays/systemd.network/rootfs/etc/systemd/network/10-persistent-net.link.ww

// overlays/debug/rootfs/warewulf/template-variables.md.ww

// overlays/host/rootfs/etc/dhcp/dhcpd.conf.ww
// overlays/host/rootfs/etc/dnsmasq.d/ww4-hosts.conf.ww
// overlays/host/rootfs/etc/exports.ww
// overlays/host/rootfs/etc/hosts.ww
// overlays/host/rootfs/etc/profile.d/ssh_setup.csh.ww
// overlays/host/rootfs/etc/profile.d/ssh_setup.sh.ww

// overlays/ignition/rootfs/etc/systemd/system/ww4-disks.target.ww
// overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww
// overlays/ignition/rootfs/etc/warewulf/ignition.json.ww

const ifcfg string = `backupFile: true
writeFile: true
Filename: ifcfg-default.conf

# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/ifcfg/rootfs/etc/sysconfig/network-scripts/ifcfg.ww
TYPE=ethernet
DEVICE=wwnet0
NAME=default
BOOTPROTO=static
DEVTIMEOUT=10
IPADDR=192.168.3.21
NETMASK=255.255.255.0
GATEWAY=192.168.3.1
HWADDR=e6:92:39:49:7b:03
TYPE=ethernet
ONBOOT=true
IPV6INIT=yes
IPV6_AUTOCONF=yes
IPV6_DEFROUTE=yes
IPV6_FAILURE_FATAL=no
backupFile: true
writeFile: true
Filename: ifcfg-secondary.conf
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/ifcfg/rootfs/etc/sysconfig/network-scripts/ifcfg.ww
TYPE=ethernet
DEVICE=wwnet1
NAME=secondary
BOOTPROTO=static
DEVTIMEOUT=10
IPADDR=192.168.3.22
NETMASK=255.255.255.0
GATEWAY=192.168.3.1
HWADDR=9a:77:29:73:14:f1
TYPE=ethernet
ONBOOT=true
IPV6INIT=yes
IPV6_AUTOCONF=yes
IPV6_DEFROUTE=yes
IPV6_FAILURE_FATAL=no
DNS1=8.8.8.8
DNS2=8.8.4.4
`

const ifcfg_network string = `backupFile: true
writeFile: true
Filename: etc/sysconfig/network
NETWORKING=yes
HOSTNAME=node1
`

const wwinit_warewulf_conf string = `backupFile: true
writeFile: true
Filename: etc/warewulf/warewulf.conf
WW_INTERNAL: 43
ipaddr: 192.168.0.1/24
netmask: 255.255.255.0
network: 192.168.0.0
warewulf:
  port: 9873
  secure: false
  update interval: 60
  autobuild overlays: true
  host overlay: true
  syslog: true
dhcp:
  enabled: false
tftp:
  enabled: false
nfs:
  enabled: false
`

const wwinit_config string = `backupFile: true
writeFile: true
Filename: warewulf/config
WWCONTAINER=
WWHOSTNAME=node1
WWROOT=initramfs
WWINIT=/sbin/init
WWIPMI_IPADDR="192.168.4.21"
WWIPMI_NETMASK="255.255.255.0"
WWIPMI_GATEWAY="192.168.4.1"
WWIPMI_USER="user"
WWIPMI_PASSWORD="password"
WWIPMI_WRITE="true"
`

const wwinit_wwclient string = `backupFile: true
writeFile: true
Filename: warewulf/init.d/80-wwclient
#!/bin/sh
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/wwinit/rootfs/warewulf/init.d/80-wwclient.ww

. /warewulf/config
# Only start if the systemd is not available
test -e /usr/lib/systemd/systemd && exit 0
echo "Starting wwclient"
nohup /tmp/ww4test-REMOVED/warewulf/wwclient >/var/log/wwclient.log 2>&1 </dev/null &
`

// const ssh_authorized_keys string = `backupFile: true
// writeFile: true
// Filename: root/.ssh/authorized_keys

// `
