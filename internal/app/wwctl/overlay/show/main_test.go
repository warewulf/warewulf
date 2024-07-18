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
    disks:
      /dev/vda:
        wipe_table: "true"
        partitions:
          scratch:
            should_exist: "true"
          swap:
            number: "1"
            size_mib: "1024"
    filesystems:
      /dev/disk/by-partlabel/scratch:
        format: btrfs
        path: /scratch
        wipe_filesystem: "true"
      /dev/disk/by-partlabel/swap:
        format: swap
        path: swap
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

	env.ImportFile(t, "var/lib/warewulf/overlays/hostname/rootfs/etc/hostname.ww", "../../../../../overlays/hostname/rootfs/etc/hostname.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/issue/rootfs/etc/issue.ww", "../../../../../overlays/issue/rootfs/etc/issue.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/resolv/rootfs/etc/resolv.conf.ww", "../../../../../overlays/resolv/rootfs/etc/resolv.conf.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/wicked/rootfs/etc/wicked/ifconfig/ifcfg.xml.ww", "../../../../../overlays/wicked/rootfs/etc/wicked/ifconfig/ifcfg.xml.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/debian.interfaces/rootfs/etc/network/interfaces.d/default.ww", "../../../../../overlays/debian.interfaces/rootfs/etc/network/interfaces.d/default.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/udev/rootfs/etc/udev/rules.d/70-ww4-netname.rules.ww", "../../../../../overlays/udev/rootfs/etc/udev/rules.d/70-ww4-netname.rules.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/systemd.network/rootfs/etc/systemd/network/10-persistent-net.link.ww", "../../../../../overlays/systemd.network/rootfs/etc/systemd/network/10-persistent-net.link.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/debug/rootfs/warewulf/template-variables.md.ww", "../../../../../overlays/debug/rootfs/warewulf/template-variables.md.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/dhcp/dhcpd.conf.ww", "../../../../../overlays/host/rootfs/etc/dhcp/dhcpd.conf.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/dnsmasq.d/ww4-hosts.conf.ww", "../../../../../overlays/host/rootfs/etc/dnsmasq.d/ww4-hosts.conf.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/exports.ww", "../../../../../overlays/host/rootfs/etc/exports.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/hosts.ww", "../../../../../overlays/host/rootfs/etc/hosts.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/profile.d/ssh_setup.csh.ww", "../../../../../overlays/host/rootfs/etc/profile.d/ssh_setup.csh.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/host/rootfs/etc/profile.d/ssh_setup.sh.ww", "../../../../../overlays/host/rootfs/etc/profile.d/ssh_setup.sh.ww")

	env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-disks.target.ww", "../../../../../overlays/ignition/rootfs/etc/systemd/system/ww4-disks.target.ww")
	env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww", "../../../../../overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww")
	// env.ImportFile(t, "var/lib/warewulf/overlays/ignition/rootfs/etc/warewulf/ignition.json.ww", "../../../../../overlays/ignition/rootfs/etc/warewulf/ignition.json.ww")

	// Some overlay templates can't be relably tested because they
	// depend on build host files. Such tests are provided as
	// examples but are left commented-out.
	tests := []struct {
		name         string
		warewulfconf string
		args         []string
		log          string
	}{
		{
			name:         "ifcfg:ifcfg.ww",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ifcfg", "etc/sysconfig/network-scripts/ifcfg.ww"},
			log:          ifcfg,
		},
		{
			name:         "ifcfg:network.ww",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ifcfg", "etc/sysconfig/network.ww"},
			log:          ifcfg_network,
		},
		{
			name:         "wwinit:warewulf.conf.ww",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "wwinit", "etc/warewulf/warewulf.conf.ww"},
			log:          wwinit_warewulf_conf,
		},
		{
			name:         "wwinit:config.ww",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "wwinit", "warewulf/config.ww"},
			log:          wwinit_config,
		},
		{
			name:         "wwinit:wwclient.ww",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "wwinit", "warewulf/init.d/80-wwclient.ww"},
			log:          wwinit_wwclient,
		},
		// {
		// 	name: "ssh_authorized_keys:authorized_keys.ww",
		//	warewulfconf: warewulfconf,
		// 	args: []string{"--render", "node1", "ssh_authorized_keys", "root/.ssh/authorized_keys.ww"},
		// 	log:  ssh_authorized_keys,
		// },
		// {
		// 	name: "syncuser:passwd.ww",
		//	warewulfconf: warewulfconf,
		// 	args: []string{"--render", "node1", "syncuser", "etc/passwd.ww"},
		// 	log:  syncuser_passwd,
		// },
		// {
		// 	name: "syncuser:group.ww",
		//	warewulfconf: warewulfconf,
		// 	args: []string{"--render", "node1", "syncuser", "etc/group.ww"},
		// 	log:  syncuser_group,
		// },
		{
			name:         "/etc/hosts",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "hosts", "etc/hosts.ww"},
			log:          hosts,
		},
		{
			name:         "ssh_host_keys:dsa pub",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_dsa_key.pub.ww"},
			log:          ssh_host_dsa_key_pub,
		},
		{
			name:         "ssh_host_keys:dsa",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_dsa_key.ww"},
			log:          ssh_host_dsa_key,
		},
		{
			name:         "ssh_host_keys:ecdsa pub",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ecdsa_key.pub.ww"},
			log:          ssh_host_ecdsa_key_pub,
		},
		{
			name:         "ssh_host_keys:ecdsa",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ecdsa_key.ww"},
			log:          ssh_host_ecdsa_key,
		},
		{
			name:         "ssh_host_keys:rsa pub",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_rsa_key.pub.ww"},
			log:          ssh_host_rsa_key_pub,
		},
		{
			name:         "ssh_host_keys:dsa",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_rsa_key.ww"},
			log:          ssh_host_rsa_key,
		},
		{
			name:         "ssh_host_keys:ed25519 pub",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ed25519_key.pub.ww"},
			log:          ssh_host_ed25519_key_pub,
		},
		{
			name:         "ssh_host_keys:ed25519",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ssh_host_keys", "etc/ssh/ssh_host_ed25519_key.ww"},
			log:          ssh_host_ed25519_key,
		},
		{
			name:         "NetworkManager:ww4-unmanaged.ww",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "NetworkManager", "etc/NetworkManager/conf.d/ww4-unmanaged.ww"},
			log:          networkmanager_unmanaged,
		},
		{
			name:         "NetworkManager:ww4-managed.ww",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "NetworkManager", "etc/NetworkManager/system-connections/ww4-managed.ww"},
			log:          networkmanager_managed,
		},
		{
			name:         "/etc/fstab",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "fstab", "etc/fstab.ww"},
			log:          fstab,
		},
		{
			name:         "/etc/hostname",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "hostname", "etc/hostname.ww"},
			log:          hostname,
		},
		{
			name:         "/etc/issue",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "issue", "etc/issue.ww"},
			log:          issue,
		},
		{
			name:         "/etc/resolv.conf",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "resolv", "etc/resolv.conf.ww"},
			log:          resolv_conf,
		},
		{
			name:         "wicked",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "wicked", "etc/wicked/ifconfig/ifcfg.xml.ww"},
			log:          wicked,
		},
		{
			name:         "debian interfaces",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "debian.interfaces", "etc/network/interfaces.d/default.ww"},
			log:          debian_interfaces,
		},
		{
			name:         "udev netnames",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "udev", "etc/udev/rules.d/70-ww4-netname.rules.ww"},
			log:          udev_netnames,
		},
		{
			name:         "systemd network links",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "systemd.network", "etc/systemd/network/10-persistent-net.link.ww"},
			log:          systemd_network_links,
		},
		{
			name:         "debug",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "debug", "warewulf/template-variables.md.ww"},
			log:          debug,
		},
		{
			name:         "host:dhcp",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "host", "host", "etc/dhcp/dhcpd.conf.ww"},
			log:          host_dhcp,
		},
		{
			name:         "host:dhcp(static)",
			warewulfconf: warewulfconf_static_dhcp,
			args:         []string{"--render", "host", "host", "etc/dhcp/dhcpd.conf.ww"},
			log:          host_dhcp_static,
		},
		{
			name:         "host:dnsmasq",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "host", "host", "etc/dnsmasq.d/ww4-hosts.conf.ww"},
			log:          host_dnsmasq,
		},
		{
			name:         "host:/etc/exports",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "host", "host", "etc/exports.ww"},
			log:          host_exports,
		},
		{
			name:         "host:/etc/hosts",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "host", "host", "etc/hosts.ww"},
			log:          host_hosts,
		},
		{
			name:         "host:ssh_setup.csh",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "host", "host", "etc/profile.d/ssh_setup.csh.ww"},
			log:          host_ssh_setup_csh,
		},
		{
			name:         "host:ssh_setup.sh",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "host", "host", "etc/profile.d/ssh_setup.sh.ww"},
			log:          host_ssh_setup_sh,
		},
		{
			name:         "ignition:ww4-disks.target",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ignition", "etc/systemd/system/ww4-disks.target.ww"},
			log:          ignition_disks,
		},
		{
			name:         "ignition:ww4-mounts",
			warewulfconf: warewulfconf,
			args:         []string{"--render", "node1", "ignition", "etc/systemd/system/ww4-mounts.ww"},
			log:          ignition_mounts,
		},
		// {
		// 	name:         "ignition:ignition.json",
		// 	warewulfconf: warewulfconf,
		// 	args:         []string{"--render", "node1", "ignition", "etc/warewulf/ignition.json.ww"},
		// 	log:          ignition_json,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env.WriteFile(t, "etc/warewulf/warewulf.conf", tt.warewulfconf)
			assert.NoError(t, config.Get().Read(env.GetPath("etc/warewulf/warewulf.conf")))

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
	host := regexp.MustCompile(`(?m)^(# |)Host:.*`)
	time := regexp.MustCompile(`(?m)^(# |)Time:.*`)
	source := regexp.MustCompile(`/ww4test-[0-9]*/`)
	hostname_, _ := os.Hostname()
	hostname := regexp.MustCompile(`(?m) ` + hostname_ + ` warewulf$`)
	buildTimeList := regexp.MustCompile(`(?m)^- BuildTime: .*`)
	buildTimeUnixList := regexp.MustCompile(`(?m)^- BuildTimeUnix: .*`)
	buildHostList := regexp.MustCompile(`(?m)^- BuildHost: .*`)

	output = input
	output = host.ReplaceAllString(output, "${1}Host:   REMOVED")
	output = time.ReplaceAllString(output, "${1}Time:   REMOVED")
	output = source.ReplaceAllString(output, "/ww4test-REMOVED/")
	output = hostname.ReplaceAllString(output, " REMOVED warewulf")
	output = buildTimeList.ReplaceAllString(output, "- BuildTime: REMOVED")
	output = buildTimeUnixList.ReplaceAllString(output, "- BuildTimeUnix: REMOVED")
	output = buildHostList.ReplaceAllString(output, "- BuildHost: REMOVED")
	return output
}

const warewulfconf string = `WW_INTERNAL: 43
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
  enabled: true
  range start: 192.168.0.100
  range end: 192.168.0.199
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
    mount: false`

const warewulfconf_static_dhcp string = `WW_INTERNAL: 43
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
  enabled: true
  template: static
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
    mount: false`

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
  enabled: true
  range start: 192.168.0.100
  range end: 192.168.0.199
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
# mounts for local file systems created with ignition in nodes.conf
# all with noauto as mounts happens with systemd units
/dev/disk/by-partlabel/scratch /scratch btrfs noauto,defaults 0 0
/dev/disk/by-partlabel/swap swap swap noauto,defaults 0 0
# nfs mounts provided in warewulf.conf
192.168.0.1:/home /home nfs defaults 0 0
`

const hostname string = `backupFile: true
writeFile: true
Filename: etc/hostname
node1
`

const issue string = `backupFile: true
writeFile: true
Filename: etc/issue
Warewulf Node:      node1
Container:          rockylinux-9
Kernelargs:         quiet crashkernel=no vga=791 net.naming-scheme=v238

Network:
    default: wwnet0
    default: 192.168.3.21/24
    default: e6:92:39:49:7b:03
    secondary: wwnet1
    secondary: 192.168.3.22/24
    secondary: 9a:77:29:73:14:f1
`

const resolv_conf string = `backupFile: true
writeFile: true
Filename: etc/resolv.conf

# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/resolv/rootfs/etc/resolv.conf.ww
nameserver 8.8.8.8
nameserver 8.8.4.4
`
const wicked string = `backupFile: true
writeFile: true
Filename: ifcfg-default.xml

<!--
This file is autogenerated by warewulf
Host:   REMOVED
Time:   REMOVED
Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/wicked/rootfs/etc/wicked/ifconfig/ifcfg.xml.ww
-->
<interface origin="static generated warewulf config">
  <name>wwnet0</name>
  <link-type>ethernet</link-type>
  <control>
    <mode>boot</mode>
  </control>
  <firewall/>
  <link/>
  <ipv4>
    <enabled>true</enabled>
    <arp-verify>true</arp-verify>
  </ipv4>
  <ipv4:static>
    <address>
      <local>192.168.3.21/24</local>
    </address>
<route>
      <nexthop>
        <gateway>192.168.3.1</gateway>
      </nexthop>
    </route>
</ipv4:static>
  <ipv6>
    <enabled>true</enabled>
    <privacy>prefer-public</privacy>
    <accept-redirects>false</accept-redirects>
  </ipv6>
</interface>
backupFile: true
writeFile: true
Filename: ifcfg-secondary.xml
<!--
This file is autogenerated by warewulf
Host:   REMOVED
Time:   REMOVED
Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/wicked/rootfs/etc/wicked/ifconfig/ifcfg.xml.ww
-->
<interface origin="static generated warewulf config">
  <name>wwnet1</name>
  <link-type>ethernet</link-type>
  <control>
    <mode>boot</mode>
  </control>
  <firewall/>
  <link/>
  <ipv4>
    <enabled>true</enabled>
    <arp-verify>true</arp-verify>
  </ipv4>
  <ipv4:static>
    <address>
      <local>192.168.3.22/24</local>
    </address>
<route>
      <nexthop>
        <gateway>192.168.3.1</gateway>
      </nexthop>
    </route>
</ipv4:static>
  <ipv6>
    <enabled>true</enabled>
    <privacy>prefer-public</privacy>
    <accept-redirects>false</accept-redirects>
  </ipv6>
</interface>
`

const debian_interfaces string = `backupFile: true
writeFile: true
Filename: default

# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/debian.interfaces/rootfs/etc/network/interfaces.d/default.ww
auto wwnet0
allow-hotplug wwnet0
iface wwnet0 inet static
  address 192.168.3.21
  netmask 255.255.255.0
  gateway 192.168.3.1
  mtu 
backupFile: true
writeFile: true
Filename: secondary
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/debian.interfaces/rootfs/etc/network/interfaces.d/default.ww
auto wwnet1
allow-hotplug wwnet1
iface wwnet1 inet static
  address 192.168.3.22
  netmask 255.255.255.0
  gateway 192.168.3.1
  mtu 
  up ifmetric wwnet1 30
`

const udev_netnames string = `backupFile: true
writeFile: true
Filename: etc/udev/rules.d/70-ww4-netname.rules
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/udev/rootfs/etc/udev/rules.d/70-ww4-netname.rules.ww

SUBSYSTEM=="net", ACTION=="add", ATTR{address}=="e6:92:39:49:7b:03", NAME="wwnet0"

SUBSYSTEM=="net", ACTION=="add", ATTR{address}=="9a:77:29:73:14:f1", NAME="wwnet1"
`

const systemd_network_links string = `backupFile: true
writeFile: true
Filename: 10-persistent-net-default.link
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/systemd.network/rootfs/etc/systemd/network/10-persistent-net.link.ww
[Match]
MACAddress=e6:92:39:49:7b:03
[Link]
Name=wwnet0
backupFile: true
writeFile: true
Filename: 10-persistent-net-secondary.link
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/systemd.network/rootfs/etc/systemd/network/10-persistent-net.link.ww
[Match]
MACAddress=9a:77:29:73:14:f1
[Link]
Name=wwnet1
`

const debug string = `backupFile: true
writeFile: true
Filename: warewulf/template-variables.md
# Warewulf template variables

This Warewulf template serves as a complete example of the variables
available to Warewulf templates. It may also be rendered against a
node to debug its apparent configuration.

` + "```" + `sh
wwctl overlay show --render $nodename debug /warewulf/template-variables.md.ww
` + "```" + `

The template data structure is defined in
` + "`internal/pkg/overlay/datastructure.go`" + `, though it also references
data from other structures.


## Node

- Id: node1
- Hostname: node1
- Comment: 
- ClusterName: 
- ContainerName: rockylinux-9
- Ipxe: default
- RuntimeOverlay: hosts,ssh_authorized_keys,syncuser
- SystemOverlay: wwinit,fstab,hostname,ssh_host_keys,issue,resolv,udev,systemd.network,ifcfg,NetworkManager,debian.interfaces,wicked,ignition
- Init: /sbin/init
- Root: initramfs
- AssetKey: 
- Discoverable: 
- Profiles: empty
- Tags: 
- TagsDel: 
- Keys: 
- Kernel:
  - Version: 
  - Override: 
  - Args: quiet crashkernel=no vga=791 net.naming-scheme=v238
- Ipmi:
  - UserName: user
  - Password: password
  - Ipaddr: 192.168.4.21
  - Netmask: 255.255.255.0
  - Port: 
  - Gateway: 192.168.4.1
  - Interface: 
  - Write: true
  - Tags: 
  - TagsDel: 
- NetDevs[default]:
  - Type: ethernet
  - OnBoot: true
  - Device: wwnet0
  - Hwaddr: e6:92:39:49:7b:03
  - Ipaddr: 192.168.3.21
  - IpCIDR: 192.168.3.21/24
  - Ipaddr6: 
  - Prefix: 
  - Netmask: 255.255.255.0
  - Gateway: 192.168.3.1
  - MTU: 
  - Primary: true
  - Default: 
  - Tags: 
  - TagsDel: 
- NetDevs[secondary]:
  - Type: ethernet
  - OnBoot: true
  - Device: wwnet1
  - Hwaddr: 9a:77:29:73:14:f1
  - Ipaddr: 192.168.3.22
  - IpCIDR: 192.168.3.22/24
  - Ipaddr6: 
  - Prefix: 
  - Netmask: 255.255.255.0
  - Gateway: 192.168.3.1
  - MTU: 
  - Primary: 
  - Default: 
  - Tags: DNS1=8.8.8.8 DNS2=8.8.4.4 
  - TagsDel: 


## Build variables

- BuildHost: REMOVED
- BuildTime: REMOVED
- BuildTimeUnix: REMOVED
- BuildSource: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/debug/rootfs/warewulf/template-variables.md.ww
- Overlay: debug


## Network

- Ipaddr: 192.168.0.1
- Ipaddr6: 
- Netmask: 255.255.255.0
- Network: 192.168.0.0
- NetworkCIDR: 192.168.0.0/24
- Ipv6: false


## Services

### DHCP

- Dhcp.Enabled: true
- Dhcp.Template: default
- Dhcp.RangeStart: 192.168.0.100
- Dhcp.RangeEnd: 192.168.0.199
- Dhcp.SystemdName: dhcpd


### NFS

- Enabled: true
- SystemdName: nfsd
- ExportsExtended[0]:
  - Path: /home
  - ExportOptions: rw,sync
  - MountOptions: defaults
  - Mount: true
- ExportsExtended[1]:
  - Path: /opt
  - ExportOptions: ro,sync,no_root_squash
  - MountOptions: defaults
  - Mount: false

### SSH
- Key types:
  - rsa
  - dsa
  - ecdsa
  - ed25519
- First key type: rsa

### Warewulf

- Port: 9873
- Secure: false
- UpdateInterval: 60
- AutobuildOverlays: true
- EnableHostOverlay: true
- Syslog: true
- DataStore: /tmp/ww4test-REMOVED/share


### Other nodes

- AllNodes[0]:
  - Id: node1
  - Comment: 
  - ClusterName: 
  - ContainerName: rockylinux-9
  - Ipxe: default
  - RuntimeOverlay: hosts
  - SystemOverlay: wwinit
  - Root: initramfs
  - Discoverable: 
  - Init: /sbin/init
  - AssetKey: 
  - Profiles: empty
  - Tags: 
  - Kernel
    - Override: 
    - Args: quiet crashkernel=no vga=791 net.naming-scheme=v238
  - Ipmi:
    - Ipaddr: 192.168.4.21
    - Netmask: 255.255.255.0
    - Port: 
    - Gateway: 192.168.4.1
    - UserName: user
    - Password: password
    - Interface: 
    - Write: true
    - Tags: 
  - NetDevs[default]:
    - Type: ethernet
    - OnBoot: true
    - Device: wwnet0
    - Hwaddr: e6:92:39:49:7b:03
    - Ipaddr: 192.168.3.21
    - Ipaddr6: 
    - IpCIDR: 
    - Prefix: 
    - Netmask: 255.255.255.0
    - Gateway: 192.168.3.1
    - MTU: 
    - Primary: true
    - Tags: 
  - NetDevs[secondary]:
    - Type: ethernet
    - OnBoot: true
    - Device: wwnet1
    - Hwaddr: 9a:77:29:73:14:f1
    - Ipaddr: 192.168.3.22
    - Ipaddr6: 
    - IpCIDR: 
    - Prefix: 
    - Netmask: 255.255.255.0
    - Gateway: 192.168.3.1
    - MTU: 
    - Primary: 
    - Tags: DNS1=8.8.8.8 DNS2=8.8.4.4 
- AllNodes[1]:
  - Id: node2
  - Comment: 
  - ClusterName: 
  - ContainerName: 
  - Ipxe: default
  - RuntimeOverlay: hosts
  - SystemOverlay: wwinit
  - Root: initramfs
  - Discoverable: 
  - Init: /sbin/init
  - AssetKey: 
  - Profiles: empty
  - Tags: 
  - Kernel
    - Override: 
    - Args: quiet crashkernel=no vga=791 net.naming-scheme=v238
  - Ipmi:
    - Ipaddr: 
    - Netmask: 
    - Port: 
    - Gateway: 
    - UserName: 
    - Password: 
    - Interface: 
    - Write: 
    - Tags: 
  - NetDevs[default]:
    - Type: ethernet
    - OnBoot: true
    - Device: wwnet0
    - Hwaddr: e6:92:39:49:7b:04
    - Ipaddr: 192.168.3.23
    - Ipaddr6: 
    - IpCIDR: 
    - Prefix: 
    - Netmask: 255.255.255.0
    - Gateway: 192.168.3.1
    - MTU: 
    - Primary: true
    - Tags: 

`

const host_dhcp string = `backupFile: true
writeFile: true
Filename: etc/dhcp/dhcpd.conf
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/host/rootfs/etc/dhcp/dhcpd.conf.ww

allow booting;
allow bootp;
ddns-update-style interim;
authoritative;

option space ipxe;

# Tell iPXE to not wait for ProxyDHCP requests to speed up boot.
option ipxe.no-pxedhcp code 176 = unsigned integer 8;
option ipxe.no-pxedhcp 1;

option space PXE;
option PXE.mtftp-ip    code 1 = ip-address;
option PXE.mtftp-cport code 2 = unsigned integer 16;
option PXE.mtftp-sport code 3 = unsigned integer 16;
option PXE.mtftp-tmout code 4 = unsigned integer 8;
option PXE.mtftp-delay code 5 = unsigned integer 8;

option architecture-type   code 93  = unsigned integer 16;
if exists user-class and option user-class = "iPXE" {
    filename "http://192.168.0.1:9873/ipxe/${mac:hexhyp}?assetkey=${asset}&uuid=${uuid}";
} else {

    if option architecture-type = 00:00 {
        filename "/warewulf/undionly.kpxe";
    }
    if option architecture-type = 00:07 {
        filename "/warewulf/ipxe-snponly-x86_64.efi";
    }
    if option architecture-type = 00:09 {
        filename "/warewulf/ipxe-snponly-x86_64.efi";
    }
    if option architecture-type = 00:0B {
        filename "/warewulf/snponly.efi";
    }
}

subnet 192.168.0.0 netmask 255.255.255.0 {
    max-lease-time 120;
    range 192.168.0.100 192.168.0.199;
    next-server 192.168.0.1;
}
`

const host_dhcp_static string = `backupFile: true
writeFile: true
Filename: etc/dhcp/dhcpd.conf
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/host/rootfs/etc/dhcp/dhcpd.conf.ww

allow booting;
allow bootp;
ddns-update-style interim;
authoritative;

option space ipxe;

# Tell iPXE to not wait for ProxyDHCP requests to speed up boot.
option ipxe.no-pxedhcp code 176 = unsigned integer 8;
option ipxe.no-pxedhcp 1;

option space PXE;
option PXE.mtftp-ip    code 1 = ip-address;
option PXE.mtftp-cport code 2 = unsigned integer 16;
option PXE.mtftp-sport code 3 = unsigned integer 16;
option PXE.mtftp-tmout code 4 = unsigned integer 8;
option PXE.mtftp-delay code 5 = unsigned integer 8;

option architecture-type   code 93  = unsigned integer 16;
if exists user-class and option user-class = "iPXE" {
    filename "http://192.168.0.1:9873/ipxe/${mac:hexhyp}?assetkey=${asset}&uuid=${uuid}";
} else {

    if option architecture-type = 00:00 {
        filename "/warewulf/undionly.kpxe";
    }
    if option architecture-type = 00:07 {
        filename "/warewulf/ipxe-snponly-x86_64.efi";
    }
    if option architecture-type = 00:09 {
        filename "/warewulf/ipxe-snponly-x86_64.efi";
    }
    if option architecture-type = 00:0B {
        filename "/warewulf/snponly.efi";
    }
}

subnet 192.168.0.0 netmask 255.255.255.0 {
    max-lease-time 120;
}
host node1-default
{
    hardware ethernet e6:92:39:49:7b:03;
    fixed-address 192.168.3.21;
    option host-name "node1";
}

host node1-secondary
{
    hardware ethernet 9a:77:29:73:14:f1;
    fixed-address 192.168.3.22;
}


host node2-default
{
    hardware ethernet e6:92:39:49:7b:04;
    fixed-address 192.168.3.23;
    option host-name "node2";
}



`

const host_dnsmasq string = `backupFile: false
writeFile: true
Filename: etc/dnsmasq.d/ww4-hosts.conf
# This file was autgenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source  /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/host/rootfs/etc/dnsmasq.d/ww4-hosts.conf.ww

# select the x86 hosts which will get the iXPE binary
dhcp-match=set:bios,option:client-arch,0   #legacy boot
dhcp-match=set:x86PC,option:client-arch, 7 #EFI x86-64
dhcp-match=set:x86PC,option:client-arch, 6 #EFI x86-64
dhcp-match=set:x86PC,option:client-arch, 9 #EFI x86-64
dhcp-match=set:aarch64,option:client-arch, 11 #EFI aarch64
dhcp-match=set:iPXE,77,"iPXE"
dhcp-userclass=set:iPXE,iPXE
dhcp-vendorclass=set:efi-http,HTTPClient:Arch:00016
dhcp-option-force=tag:efi-http,60,HTTPClient
# for http boot always use shim/grub
dhcp-boot=tag:efi-http,"http://192.168.0.1:9873/efiboot/shim.efi"
dhcp-boot=tag:x86PC,"/warewulf/ipxe-snponly-x86_64.efi"
dhcp-boot=tag:aarch64,"/warewulf/arm64-efi/snponly.efi"
# iPXE binary will get the following configuration file
dhcp-boot=tag:iPXE,"http://192.168.0.1:9873/ipxe/${mac:hexhyp}?assetkey=${asset}&uuid=${uuid}"
dhcp-no-override
# define the the range
dhcp-range=192.168.0.100,192.168.0.199,255.255.255.0,6h
                                   
dhcp-host=e6:92:39:49:7b:03,set:warewulf,node1,192.168.3.21,infinite                 
dhcp-host=9a:77:29:73:14:f1,set:warewulf,node1,192.168.3.22,infinite                                   
dhcp-host=e6:92:39:49:7b:04,set:warewulf,node2,192.168.3.23,infinite
`

const host_exports string = `backupFile: true
writeFile: true
Filename: etc/exports

# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/host/rootfs/etc/exports.ww
/home 192.168.0.0/255.255.255.0(rw,sync)
/opt 192.168.0.0/255.255.255.0(ro,sync,no_root_squash)

`

const host_hosts string = `backupFile: true
writeFile: true
Filename: etc/hosts
testing
# Do not edit after this line
# This block is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/host/rootfs/etc/hosts.ww


# Warewulf Server
192.168.0.1 rocky warewulf                  
# Entry for node1                 
192.168.3.21 node1 node1-default node1-wwnet0                 
192.168.3.22  node1-secondary node1-wwnet1                  
# Entry for node2                 
192.168.3.23 node2 node2-default node2-wwnet0
`

const host_ssh_setup_csh = `backupFile: true
writeFile: true
Filename: etc/profile.d/ssh_setup.csh
#!/bin/csh

## Automatically configure SSH keys for a user on C SHell login
## Copy this file to /etc/profile.d along with ssh_setup.sh

` + "set _UID=`id -u`" + `
if ( ( $_UID > 500 || $_UID == 0 ) && ( ! -f "$HOME/.ssh/config" && ! -f "$HOME/.ssh/cluster" ) ) then
    echo "Configuring SSH for cluster access"
    install -d -m 700 $HOME/.ssh
    ssh-keygen -t rsa -f $HOME/.ssh/cluster -N '' -C "Warewulf Cluster key" >& /dev/null
    cat $HOME/.ssh/cluster.pub >>! $HOME/.ssh/authorized_keys
    chmod 0600 $HOME/.ssh/authorized_keys

    touch $HOME/.ssh/config
    echo -n "# Added by Warewulf " >>! $HOME/.ssh/config
    (date +%!Y(MISSING)-%!m(MISSING)-%!d(MISSING) >> $HOME/.ssh/config) >& /dev/null
    echo "Host *" >> $HOME/.ssh/config
    echo "   IdentityFile ~/.ssh/cluster" >> $HOME/.ssh/config
    echo "   StrictHostKeyChecking=no" >> $HOME/.ssh/config
    chmod 0600 $HOME/.ssh/config
endif
`

const host_ssh_setup_sh = `backupFile: true
writeFile: true
Filename: etc/profile.d/ssh_setup.sh
#!/bin/sh
##
## Copyright (c) 2001-2003 Gregory M. Kurtzer
##
## Copyright (c) 2003-2012, The Regents of the University of California,
## through Lawrence Berkeley National Laboratory (subject to receipt of any
## required approvals from the U.S. Dept. of Energy).  All rights reserved.
##
## Copied from https://github.com/warewulf/warewulf3/blob/master/cluster/bin/cluster-env

## Automatically configure SSH keys for a user on login
## Copy this file to /etc/profile.d

` + "_UID=`id -u`" + `
if [ $_UID -ge 500 -o $_UID -eq 0 ] && [ ! -f "$HOME/.ssh/config" -a ! -f "$HOME/.ssh/cluster" ]; then
    echo "Configuring SSH for cluster access"
    install -d -m 700 $HOME/.ssh
    ssh-keygen -t rsa -f $HOME/.ssh/cluster -N '' -C "Warewulf Cluster key" > /dev/null 2>&1
    cat $HOME/.ssh/cluster.pub >> $HOME/.ssh/authorized_keys
    chmod 0600 $HOME/.ssh/authorized_keys

    echo "# Added by Warewulf  ` + "`date +%!Y(MISSING)-%!m(MISSING)-%!d(MISSING) 2>/dev/null`" + `" >> $HOME/.ssh/config
    echo "Host *" >> $HOME/.ssh/config
    echo "   IdentityFile ~/.ssh/cluster" >> $HOME/.ssh/config
    echo "   StrictHostKeyChecking=no" >> $HOME/.ssh/config
    chmod 0600 $HOME/.ssh/config
fi
`

const ignition_disks string = `backupFile: true
writeFile: true
Filename: etc/systemd/system/ww4-disks.target
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-disks.target.ww
[Unit]
Description=mount ww4 disks
# make sure that the disks are available
Requires=ignition-ww4-disks.service
After=ignition-ww4-disks.service
Requisite=ignition-ww4-disks.service
# Get the mounts
Wants=scratch.mount
Wants=dev-disk-by\x2dpartlabel-swap.swap
`

const ignition_mounts string = `backupFile: true
writeFile: true
Filename: scratch.mount

# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww
[Unit]
ConditionPathExists=/warewulf/ignition.json
Before=local-fs.target
After=ignition-ww4-disks.service
[Mount]
Where=/scratch
What=/dev/disk/by-partlabel/scratch
Type=btrfs
[Install]
RequiredBy=local-fs.target
backupFile: true
writeFile: true
Filename: dev-disk-by\x2dpartlabel-swap.swap
# This file is autogenerated by warewulf
# Host:   REMOVED
# Time:   REMOVED
# Source: /tmp/ww4test-REMOVED/var/lib/warewulf/overlays/ignition/rootfs/etc/systemd/system/ww4-mounts.ww
[Unit]
ConditionPathExists=/warewulf/ignition.json
After=ignition-ww4-disks.service
Before=swap.target
[Swap]
What=/dev/disk/by-partlabel/swap
[Install]
RequiredBy=swap.target
`

// const ignition_json string = `backupFile: true
// writeFile: true
// Filename: etc/warewulf/ignition.json
// {"ignition":{"version":"3.1.0"},"storage":{"disks":[{"device":"/dev/vda","partitions":[{"label":"scratch","shouldExist":true,"wipePartitionEntry":false},{"label":"swap","number":1,"shouldExist":false,"sizeMiB":1024,"wipePartitionEntry":false}],"wipeTable":true}],"filesystems":[{"device":"/dev/disk/by-partlabel/scratch","format":"btrfs","path":"/scratch","wipeFilesystem":true},{"device":"/dev/disk/by-partlabel/swap","format":"swap","path":"swap","wipeFilesystem":false}]}}`
