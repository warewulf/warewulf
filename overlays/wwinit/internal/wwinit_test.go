package wwinit

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_wwinitOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.ImportFile("etc/warewulf/warewulf.conf", "warewulf.conf")
	env.ImportFile("etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile("var/lib/warewulf/overlays/wwinit/rootfs/etc/warewulf/warewulf.conf.ww", "../rootfs/etc/warewulf/warewulf.conf.ww")
	env.ImportFile("var/lib/warewulf/overlays/wwinit/rootfs/warewulf/config.ww", "../rootfs/warewulf/config.ww")
	env.ImportFile("var/lib/warewulf/overlays/wwinit/rootfs/warewulf/init.d/50-ipmi.ww", "../rootfs/warewulf/init.d/50-ipmi.ww")

	tests := []struct {
		name string
		args []string
		log  string
	}{
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
			name: "wwinit:50-ipmi.ww",
			args: []string{"--render", "node1", "wwinit", "warewulf/init.d/50-ipmi.ww"},
			log:  wwinit_50_ipmi,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := show.GetCommand()
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
			assert.Equal(t, tt.log, logbuf.String())
		})
	}
}

const wwinit_warewulf_conf string = `backupFile: true
writeFile: true
Filename: etc/warewulf/warewulf.conf
ipaddr: 192.168.0.1/24
netmask: 255.255.255.0
network: 192.168.0.0
warewulf:
  port: 9873
  secure: false
  update interval: 60
  autobuild overlays: true
  host overlay: true
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
  - path: /opt
    export options: ro,sync,no_root_squash
`

const wwinit_config string = `backupFile: true
writeFile: true
Filename: warewulf/config
WWIMAGE=rockylinux-9
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

const wwinit_50_ipmi string = `backupFile: true
writeFile: true
Filename: warewulf/init.d/50-ipmi
#!/bin/sh

. /warewulf/config

export PATH=/usr/bin:/bin:/usr/sbin:/sbin

echo "Warewulf prescript: IPMI"
echo

if [ "$WWIPMI_WRITE" != "true" ]; then
    echo "IPMI write not configured: skipping"
    exit
fi

echo "Loading IPMI kernel modules..."
modprobe ipmi_si ipmi_ssif ipmi_devintf ipmi_msghandler || (
    echo "Unable to load IPMI kernel modules: skipping IPMI configuration"
    exit
)

if [ ! -e /dev/ipmi0 ]; then
    echo "/dev/ipmi0 does not exist; creating..."
    sleep 1
    ipmi_dev=$(grep ipmidev /proc/devices | awk '{ print $1 }')
    mknod -m 0666 /dev/ipmi0 c "$ipmi_dev" 0
fi

command -v ipmitool >/dev/null 2>&1 || (
    echo "ipmitool is not available: skipping IPMI configuration"
    exit
)

lan_info="$(ipmitool lan print 1)"

if [ -n "$WWIPMI_VLAN" ]; then
    prev_vlan=$(echo "$lan_info" | grep "^802.1q VLAN ID *:" | awk -F': ' '{print $2 }')
    if [ "$prev_vlan" == "$WWIPMI_VLAN" ] || [ "$prev_vlan" = "Disabled" -a "$WWIPMI_VLAN" = off ]; then
        echo "IPMI VLAN: $WWIPMI_VLAN"
    else
        echo "IPMI VLAN: $prev_vlan -> $WWIPMI_VLAN"
        ipmitool lan set 1 vlan id "$WWIPMI_VLAN"
    fi
fi

if [ -n "$WWIPMI_IPADDR" ]; then
    prev_ip=$(echo "$lan_info" | grep "^IP Address *:" | awk -F': ' '{ print $2 }')
    if [ "$prev_ip" != "$WWIPMI_IPADDR" ]; then
        echo "IPMI IP address: $prev_ip -> $WWIPMI_IPADDR"
        ipmitool lan set 1 ipsrc static
        ipmitool lan set 1 ipaddr "$WWIPMI_IPADDR"
        ipmitool lan set 1 access on
    else
        echo "IPMI IP address: $WWIPMI_IPADDR"
    fi
fi

if [ -n "$WWIPMI_NETMASK" ]; then
    prev_netmask=$(echo "$lan_info" | grep "^Subnet Mask *:" | awk -F': ' '{ print $2 }')
    if [ "$prev_netmask" != "$WWIPMI_NETMASK" ]; then
        echo "IPMI netmask: $prev_netmask -> $WWIPMI_NETMASK"
        ipmitool lan set 1 netmask $WWIPMI_NETMASK
    else
        echo "IPMI netmask: $WWIPMI_NETMASK"
    fi
fi

if [ -n "$WWIPMI_GATEWAY" ]; then
    prev_gateway=$(echo "$lan_info" | grep "^Default Gateway IP *:" | awk -F': ' '{ print $2 }')
    if [ "$prev_gateway" != "$WWIPMI_GATEWAY" ]; then
        echo "IPMI gateway: $prev_gateway -> $WWIPMI_GATEWAY"
        ipmitool lan set 1 defgw ipaddr "$WWIPMI_GATEWAY"
    else
        echo "IPMI gateway: $WWIPMI_GATEWAY"
    fi
fi

if [ -n "$WWIPMI_USER" ]; then
    prev_user=$(ipmitool -c user list 1 | awk -F, '{ if ($1 == 2) { print $2; exit } }')
    if [ "$prev_user" != "$WWIPMI_USER" ]; then
        ipmitool user set name 2 "$WWIPMI_USER"
        ipmitool user priv 2 4 1
        ipmitool user enable 2
        echo "IPMI username: $prev_user -> $WWIPMI_USER"
    else
        echo "IPMI username: $WWIPMI_USER"
    fi
fi

if [ -n "$WWIPMI_PASSWORD" ]; then
    ipmitool user test 2 20 "$WWIPMI_PASSWORD" >/dev/null || ipmitool user test 2 16 "$WWIPMI_PASSWORD" >/dev/null
    if [ $? -ne 0 ]; then
        ipmitool user set password 2 "$WWIPMI_PASSWORD"
        ipmitool user priv 2 4 1
        ipmitool user enable 2
        echo "IPMI password: [updated]"
    else
        echo "IPMI password: [unchanged]"
    fi
fi

echo "Configuring Serial over LAN..."
ipmitool channel setaccess 1 2 link=on ipmi=on callin=on privilege=4
ipmitool sol set force-encryption true 1
ipmitool sol set force-authentication true 1
ipmitool sol set privilege-level admin 1
ipmitool sol payload enable 1 2
ipmitool sol set enabled true 1 1
speed=115.2
ipmitool sol set non-volatile-bit-rate $speed 1
ipmitool sol set volatile-bit-rate $speed 1
`
