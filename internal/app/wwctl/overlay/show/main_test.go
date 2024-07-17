package show

import (
	"bytes"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	"github.com/warewulf/warewulf/internal/pkg/config"
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
  enabled: true
  export paths:
  - path: /home
    export options: rw,sync
    mount options: defaults
    mount: true
  - path: /opt
    export options: ro,sync,no_root_squash
    mount options: defaults
    mount: false`)
	assert.NoError(t, config.Get().Read(env.GetPath("etc/warewulf/warewulf.conf")))
	env.WriteFile(t, "etc/warewulf/nodes.conf",
		`WW_INTERNAL: 45
nodeprofiles:
  empty: {}
nodes:
  node1:
    profiles:
    - empty
    container name: rockylinux-9
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
  node2:
    profiles:
    - empty
    network devices:
      default:
        device: wwnet0
        hwaddr: e6:92:39:49:7b:04
        ipaddr: 192.168.3.23
        netmask: 255.255.255.0
        gateway: 192.168.3.1
`)
	env.ImportFile(t, "var/lib/warewulf/overlays/ifcfg/rootfs/etc/sysconfig/network-scripts/ifcfg.ww", "../../../../../overlays/ifcfg/rootfs/etc/sysconfig/network-scripts/ifcfg.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ifcfg/rootfs/etc/sysconfig/network.ww", "../../../../../overlays/ifcfg/rootfs/etc/sysconfig/network.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/wwinit/rootfs/etc/warewulf/warewulf.conf.ww", "../../../../../overlays/wwinit/rootfs/etc/warewulf/warewulf.conf.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/wwinit/rootfs/warewulf/config.ww", "../../../../../overlays/wwinit/rootfs/warewulf/config.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/wwinit/rootfs/warewulf/init.d/80-wwclient.ww", "../../../../../overlays/wwinit/rootfs/warewulf/init.d/80-wwclient.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/ssh_authorized_keys/rootfs/root/.ssh/authorized_keys.ww", "../../../../../overlays/ssh_authorized_keys/rootfs/root/.ssh/authorized_keys.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/syncuser/rootfs/etc/passwd.ww", "../../../../../overlays/syncuser/rootfs/etc/passwd.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/syncuser/rootfs/etc/group.ww", "../../../../../overlays/syncuser/rootfs/etc/group.ww")
	// env.WriteFile(t, "var/lib/warewulf/chroots/rockylinux-9/rootfs/etc/passwd", `root:x:0:0:root:/root:/bin/bash`)
	// env.WriteFile(t, "var/lib/warewulf/chroots/rockylinux-9/rootfs/etc/group", `root:x:0:`)

	env.ImportFile(t, "var/lib/warewulf/overlays/hosts/rootfs/etc/hosts.ww", "../../../../../overlays/hosts/rootfs/etc/hosts.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_dsa_key.pub.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_dsa_key.pub.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_dsa_key.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_dsa_key.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.pub.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.pub.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.pub.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.pub.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_rsa_key.pub.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_rsa_key.pub.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_rsa_key.ww", "../../../../../overlays/ssh_host_keys/rootfs/etc/ssh/ssh_host_rsa_key.ww")
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_dsa_key.pub", `dsa pubkey sentinel`)
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_dsa_key", `dsa key sentinel`)
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_ecdsa_key.pub", `ecdsa pubkey sentinel`)
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_ecdsa_key", `ecdsa key sentinel`)
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_ed25519_key.pub", `ed25519 pubkey sentinel`)
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_ed25519_key", `ed25519 key sentinel`)
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_rsa_key.pub", `rsa pubkey sentinel`)
	env.WriteFile(t, "etc/warewulf/keys/ssh_host_rsa_key", `rsa key sentinel`)

	env.ImportFile(t, "var/lib/warewulf/overlays/NetworkManager/rootfs/etc/NetworkManager/conf.d/ww4-unmanaged.ww", "../../../../../overlays/NetworkManager/rootfs/etc/NetworkManager/conf.d/ww4-unmanaged.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/NetworkManager/rootfs/etc/NetworkManager/system-connections/ww4-managed.ww", "../../../../../overlays/NetworkManager/rootfs/etc/NetworkManager/system-connections/ww4-managed.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/fstab/rootfs/etc/fstab.ww", "../../../../../overlays/fstab/rootfs/etc/fstab.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/hostname/rootfs/etc/hostname.ww", "../../../../../overlays/hostname/rootfs/etc/hostname.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/issue/rootfs/etc/issue.ww", "../../../../../overlays/issue/rootfs/etc/issue.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/resolv/rootfs/etc/resolv.conf.ww", "../../../../../overlays/resolv/rootfs/etc/resolv.conf.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/wicked/rootfs/etc/wicked/ifconfig/ifcfg.xml.ww", "../../../../../overlays/wicked/rootfs/etc/wicked/ifconfig/ifcfg.xml.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/debian.interfaces/rootfs/etc/network/interfaces.d/default.ww", "../../../../../overlays/debian.interfaces/rootfs/etc/network/interfaces.d/default.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/udev/rootfs/etc/udev/rules.d/70-ww4-netname.rules.ww", "../../../../../overlays/udev/rootfs/etc/udev/rules.d/70-ww4-netname.rules.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/systemd.network/rootfs/etc/systemd/network/10-persistent-net.link.ww", "../../../../../overlays/systemd.network/rootfs/etc/systemd/network/10-persistent-net.link.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/debug/rootfs/warewulf/template-variables.md.ww", "../../../../../overlays/debug/rootfs/warewulf/template-variables.md.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/dhcp/dhcpd.conf.ww", "../../../../../overlays/host/rootfs/etc/dhcp/dhcpd.conf.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/dnsmasq.d/ww4-hosts.conf.ww", "../../../../../overlays/host/rootfs/etc/dnsmasq.d/ww4-hosts.conf.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/exports.ww", "../../../../../overlays/host/rootfs/etc/exports.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/hosts.ww", "../../../../../overlays/host/rootfs/etc/hosts.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/profile.d/ssh_setup.csh.ww", "../../../../../overlays/host/rootfs/etc/profile.d/ssh_setup.csh.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/profile.d/ssh_setup.sh.ww", "../../../../../overlays/host/rootfs/etc/profile.d/ssh_setup.sh.ww")

	// env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-disks.target.ww", "../../../../../overlays/ignition/rootfs/etc/systemd/system/ww4-disks.target.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww", "../../../../../overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/warewulf/ignition.json.ww", "../../../../../overlays/ignition/rootfs/etc/warewulf/ignition.json.ww")

	// Some overlay templates can't be relably tested because they
	// depend on build host files. Such tests are provided as
	// examples but are left commented-out.
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
		// {
		// 	name: "syncuser:passwd.ww",
		// 	args: []string{"--render", "node1", "syncuser", "etc/passwd.ww"},
		// 	log:  syncuser_passwd,
		// },
		// {
		// 	name: "syncuser:group.ww",
		// 	args: []string{"--render", "node1", "syncuser", "etc/group.ww"},
		// 	log:  syncuser_group,
		// },
		{
			name: "hosts:hosts.ww",
			args: []string{"--render", "node1", "hosts", "etc/hosts.ww"},
			log:  hosts,
		},
		{
			name: "ssh_host_keys:dsa pub",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_dsa_key.pub.ww"},
			log:  ssh_host_dsa_key_pub,
		},
		{
			name: "ssh_host_keys:dsa",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_dsa_key.ww"},
			log:  ssh_host_dsa_key,
		},
		{
			name: "ssh_host_keys:ecdsa pub",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ecdsa_key.pub.ww"},
			log:  ssh_host_ecdsa_key_pub,
		},
		{
			name: "ssh_host_keys:ecdsa",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ecdsa_key.ww"},
			log:  ssh_host_ecdsa_key,
		},
		{
			name: "ssh_host_keys:rsa pub",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_rsa_key.pub.ww"},
			log:  ssh_host_rsa_key_pub,
		},
		{
			name: "ssh_host_keys:dsa",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_rsa_key.ww"},
			log:  ssh_host_rsa_key,
		},
		{
			name: "ssh_host_keys:ed25519 pub",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ed25519_key.pub.ww"},
			log:  ssh_host_ed25519_key_pub,
		},
		{
			name: "ssh_host_keys:ed25519",
			args: []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ed25519_key.ww"},
			log:  ssh_host_ed25519_key,
		},
		{
			name: "NetworkManager:ww4-unmanaged.ww",
			args: []string{"--render", "node1", "NetworkManager", "etc/NetworkManager/conf.d/ww4-unmanaged.ww"},
			log:  networkmanager_unmanaged,
		},
		{
			name: "NetworkManager:ww4-managed.ww",
			args: []string{"--render", "node1", "NetworkManager", "etc/NetworkManager/system-connections/ww4-managed.ww"},
			log:  networkmanager_managed,
		},
		{
			name: "fstab",
			args: []string{"--render", "node1", "fstab", "etc/fstab.ww"},
			log:  fstab,
		},
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
			assert.Equal(t, tt.log, removeHostInfo(logbuf.String()))
		})
	}
}

func removeHostInfo(input string) (output string) {
	host := regexp.MustCompile(`# Host:.*`)
	time := regexp.MustCompile(`# Time:.*`)
	source := regexp.MustCompile(`/ww4test-[0-9]*/`)
	hostname_, _ := os.Hostname()
	hostname := regexp.MustCompile(`(?m) ` + hostname_ + ` warewulf$`)

	output = input
	output = host.ReplaceAllString(output, "# Host:   REMOVED")
	output = time.ReplaceAllString(output, "# Time:   REMOVED")
	output = source.ReplaceAllString(output, "/ww4test-REMOVED/")
	output = hostname.ReplaceAllString(output, " REMOVED warewulf")
	return output
}

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
  enabled: true
  export paths:
  - path: /home
    export options: rw,sync
    mount options: defaults
    mount: true
  - path: /opt
    export options: ro,sync,no_root_squash
    mount options: defaults
    mount: false
`

const wwinit_config string = `backupFile: true
writeFile: true
Filename: warewulf/config
WWCONTAINER=rockylinux-9
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
//
// `

// const syncuser_passwd string = `backupFile: true
// writeFile: true
// Filename: etc/passwd
// # Uncomment the following line to enable passwordless root login
// # root::0:0:root:/root:/bin/bash
// root:x:0:0:root:/root:/bin/bash
// `

// const syncuser_group string = `backupFile: true
// writeFile: true
// Filename: etc/group
// root:x:0:
// `

const hosts string = `backupFile: true
writeFile: true
Filename: etc/hosts
127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4
::1         localhost localhost.localdomain localhost6 localhost6.localdomain6

# Warewulf Server
192.168.0.1 REMOVED warewulf
# Entry for node1
192.168.3.21 node1 node1-default node1-wwnet0
192.168.3.22  node1-secondary node1-wwnet1
# Entry for node2
192.168.3.23 node2 node2-default node2-wwnet0
`

const ssh_host_dsa_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_dsa_key.pub
dsa pubkey sentinel
`

const ssh_host_dsa_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_dsa_key
dsa key sentinel
`

const ssh_host_ecdsa_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ecdsa_key.pub
ecdsa pubkey sentinel
`

const ssh_host_ecdsa_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ecdsa_key
ecdsa key sentinel
`

const ssh_host_ed25519_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ed25519_key.pub
ed25519 pubkey sentinel
`

const ssh_host_ed25519_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ed25519_key
ed25519 key sentinel
`

const ssh_host_rsa_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_rsa_key.pub
rsa pubkey sentinel
`

const ssh_host_rsa_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_rsa_key
rsa key sentinel
`

const networkmanager_unmanaged string = `backupFile: true
writeFile: true
Filename: warewulf-unmanaged.conf
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/NetworkManager/rootfs/etc/NetworkManager/conf.d/ww4-unmanaged.ww
[main]
plugins=keyfile

[keyfile]
unmanaged-devices=except:interface-name:wwnet0,except:interface-name:wwnet1,
`

const networkmanager_managed string = `backupFile: true
writeFile: true
Filename: warewulf-default.conf
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/NetworkManager/rootfs/etc/NetworkManager/system-connections/ww4-managed.ww
[connection]
id=default
interface-name=wwnet0

type=ethernet
autoconnect=true
[ethernet]
mac-address=e6:92:39:49:7b:03
# bond
[ipv4]
address=192.168.3.21/24
gateway=192.168.3.1
method=manual

[ipv6]
addr-gen-mode=stable-privacy
method=ignore
never-default=true
backupFile: true
writeFile: true
Filename: warewulf-secondary.conf
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/NetworkManager/rootfs/etc/NetworkManager/system-connections/ww4-managed.ww
[connection]
id=secondary
interface-name=wwnet1

type=ethernet
autoconnect=true
[ethernet]
mac-address=9a:77:29:73:14:f1
# bond
[ipv4]
address=192.168.3.22/24
gateway=192.168.3.1
method=manual
dns=8.8.8.8;8.8.4.4;

[ipv6]
addr-gen-mode=stable-privacy
method=ignore
never-default=true
`

const fstab string = `backupFile: true
writeFile: true
Filename: etc/fstab
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/fstab/rootfs/etc/fstab.ww
rootfs / tmpfs defaults 0 0
devpts /dev/pts devpts gid=5,mode=620 0 0
tmpfs /run/shm tmpfs defaults 0 0
sysfs /sys sysfs defaults 0 0
proc /proc proc defaults 0 0
# nfs mounts provided in warewulf.conf
192.168.0.1:/home /home nfs defaults 0 0
`
