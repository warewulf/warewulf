package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		input  string
		result string
	}{
		"default": {
			input: ``,
			result: `
warewulf:
  autobuild overlays: true
  grubboot: false
  host overlay: true
  port: 9873
  secure: true
  update interval: 60
nfs:
  enabled: true
  systemd name: nfsd
dhcp:
  enabled: true
  systemd name: dhcpd
  template: default
image mounts:
- dest: /etc/resolv.conf
  source: /etc/resolv.conf
ssh:
  key types:
  - ed25519
  - ecdsa
  - rsa
  - dsa
tftp:
  enabled: true
  ipxe:
    "00:00": undionly.kpxe
    "00:07": ipxe-snponly-x86_64.efi
    "00:09": ipxe-snponly-x86_64.efi
    "00:0B": arm64-efi/snponly.efi
  systemd name: tftp
api:
  enabled: false
  allowed subnets:
  - 127.0.0.0/8
  - ::1/128
`,
		},
		"cidr": {
			input: `
ipaddr: 192.168.0.1/24
`,
			result: `
ipaddr: 192.168.0.1
network: 192.168.0.0
netmask: 255.255.255.0
warewulf:
  autobuild overlays: true
  grubboot: false
  host overlay: true
  port: 9873
  secure: true
  update interval: 60
nfs:
  enabled: true
  systemd name: nfsd
dhcp:
  enabled: true
  systemd name: dhcpd
  template: default
image mounts:
- dest: /etc/resolv.conf
  source: /etc/resolv.conf
ssh:
  key types:
  - ed25519
  - ecdsa
  - rsa
  - dsa
tftp:
  enabled: true
  ipxe:
    "00:00": undionly.kpxe
    "00:07": ipxe-snponly-x86_64.efi
    "00:09": ipxe-snponly-x86_64.efi
    "00:0B": arm64-efi/snponly.efi
  systemd name: tftp
api:
  enabled: false
  allowed subnets:
  - 127.0.0.0/8
  - ::1/128
`,
		},
		"cidr with conflicts": {
			input: `
ipaddr: 192.168.1.1/24
network: 192.168.0.0
netmask: 255.255.0.0
`,
			result: `
ipaddr: 192.168.1.1
network: 192.168.0.0
netmask: 255.255.0.0
warewulf:
  autobuild overlays: true
  grubboot: false
  host overlay: true
  port: 9873
  secure: true
  update interval: 60
nfs:
  enabled: true
  systemd name: nfsd
dhcp:
  enabled: true
  systemd name: dhcpd
  template: default
image mounts:
- dest: /etc/resolv.conf
  source: /etc/resolv.conf
ssh:
  key types:
  - ed25519
  - ecdsa
  - rsa
  - dsa
tftp:
  enabled: true
  ipxe:
    "00:00": undionly.kpxe
    "00:07": ipxe-snponly-x86_64.efi
    "00:09": ipxe-snponly-x86_64.efi
    "00:0B": arm64-efi/snponly.efi
  systemd name: tftp
api:
  enabled: false
  allowed subnets:
  - 127.0.0.0/8
  - ::1/128
`,
		},
		"ipv6 cidr": {
			input: `
ipaddr6: "2001:db8::1/64"
`,
			result: `
ipaddr6: "2001:db8::1"
warewulf:
  autobuild overlays: true
  grubboot: false
  host overlay: true
  port: 9873
  secure: true
  update interval: 60
nfs:
  enabled: true
  systemd name: nfsd
dhcp:
  enabled: true
  systemd name: dhcpd
  template: default
image mounts:
- dest: /etc/resolv.conf
  source: /etc/resolv.conf
ssh:
  key types:
  - ed25519
  - ecdsa
  - rsa
  - dsa
tftp:
  enabled: true
  ipxe:
    "00:00": undionly.kpxe
    "00:07": ipxe-snponly-x86_64.efi
    "00:09": ipxe-snponly-x86_64.efi
    "00:0B": arm64-efi/snponly.efi
  systemd name: tftp
api:
  enabled: false
  allowed subnets:
  - 127.0.0.0/8
  - ::1/128
`,
		},
		"example": {
			input: `
ipaddr: 192.168.200.1
netmask: 255.255.255.0
network: 192.168.200.0
warewulf:
  port: 9873
  secure: false
  update interval: 60
  autobuild overlays: true
  host overlay: true
dhcp:
  enabled: true
  range start: 192.168.200.50
  range end: 192.168.200.99
  systemd name: dhcpd
tftp:
  enabled: true
  systemd name: tftp
nfs:
  enabled: true
  export paths:
  - path: /home
    export options: rw,sync
  - path: /opt
    export options: ro,sync,no_root_squash
  systemd name: nfs-server
image mounts:
  - source: /etc/resolv.conf
    dest: /etc/resolv.conf
    readonly: true
`,
			result: `
ipaddr: 192.168.200.1
netmask: 255.255.255.0
network: 192.168.200.0
warewulf:
  autobuild overlays: true
  grubboot: false
  host overlay: true
  port: 9873
  secure: false
  update interval: 60
nfs:
  enabled: true
  systemd name: nfs-server
  export paths:
  - path: /home
    export options: rw,sync
  - path: /opt
    export options: ro,sync,no_root_squash
dhcp:
  enabled: true
  systemd name: dhcpd
  template: default
  range end: 192.168.200.99
  range start: 192.168.200.50
image mounts:
- dest: /etc/resolv.conf
  readonly: true
  source: /etc/resolv.conf
ssh:
  key types:
  - ed25519
  - ecdsa
  - rsa
  - dsa
tftp:
  enabled: true
  ipxe:
    "00:00": undionly.kpxe
    "00:07": ipxe-snponly-x86_64.efi
    "00:09": ipxe-snponly-x86_64.efi
    "00:0B": arm64-efi/snponly.efi
  systemd name: tftp
api:
  enabled: false
  allowed subnets:
  - 127.0.0.0/8
  - ::1/128
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conf := New()
			err := conf.Parse([]byte(tt.input), false)
			assert.NoError(t, err)
			removePaths(conf)
			result, err := conf.Dump()
			assert.NoError(t, err)
			assert.YAMLEq(t, tt.result, string(result))
		})
	}
}

// These paths are subject to change based on the environment the test is being
// run in.
func removePaths(conf *WarewulfYaml) {
	conf.Paths = nil
	conf.TFTP.TftpRoot = ""
}

func TestInitializedFromFile(t *testing.T) {
	example_warewulf_conf := ""
	tempWarewulfConf, warewulfConfErr := os.CreateTemp("", "warewulf.conf-")
	assert.NoError(t, warewulfConfErr)
	defer os.Remove(tempWarewulfConf.Name())
	_, warewulfConfErr = tempWarewulfConf.Write([]byte(example_warewulf_conf))
	assert.NoError(t, warewulfConfErr)
	assert.NoError(t, tempWarewulfConf.Sync())

	conf := New()
	assert.False(t, conf.InitializedFromFile())
	assert.NoError(t, conf.Read(tempWarewulfConf.Name(), false))
	assert.True(t, conf.InitializedFromFile())
	assert.Equal(t, conf.GetWarewulfConf(), tempWarewulfConf.Name())
}

func TestCache(t *testing.T) {
	confOrig := New()
	confCached := Get()

	assert.Equal(t, 9873, confOrig.Warewulf.Port)
	assert.Equal(t, 9873, confCached.Warewulf.Port)

	confOrig.Warewulf.Port = 9999
	assert.Equal(t, 9999, confCached.Warewulf.Port)
	assert.Equal(t, 9999, Get().Warewulf.Port)

	New()
	assert.NotEqual(t, 9999, Get().Warewulf.Port)
}

func TestIpCIDR(t *testing.T) {
	tests := map[string]struct {
		ipaddr  string
		netmask string
		cidr    string
	}{
		"blank": {
			ipaddr:  "",
			netmask: "",
			cidr:    "",
		},
		"ip only": {
			ipaddr:  "192.168.0.1",
			netmask: "",
			cidr:    "",
		},
		"netmask only": {
			ipaddr:  "",
			netmask: "255.255.255.0",
			cidr:    "",
		},
		"full": {
			ipaddr:  "192.168.0.1",
			netmask: "255.255.255.0",
			cidr:    "192.168.0.1/24",
		},
		"invalid ip": {
			ipaddr:  "asdf",
			netmask: "255.255.255.0",
			cidr:    "",
		},
		"invalid netmask": {
			ipaddr:  "192.168.0.1",
			netmask: "asdf",
			cidr:    "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conf := New()
			conf.Ipaddr = tt.ipaddr
			conf.Netmask = tt.netmask
			assert.Equal(t, tt.cidr, conf.IpCIDR())
		})
	}
}

func TestNetworkCIDR(t *testing.T) {
	tests := map[string]struct {
		network string
		netmask string
		cidr    string
	}{
		"blank": {
			network: "",
			netmask: "",
			cidr:    "",
		},
		"network only": {
			network: "192.168.0.0",
			netmask: "",
			cidr:    "",
		},
		"netmask only": {
			network: "",
			netmask: "255.255.255.0",
			cidr:    "",
		},
		"full": {
			network: "192.168.0.0",
			netmask: "255.255.255.0",
			cidr:    "192.168.0.0/24",
		},
		"invalid network": {
			network: "asdf",
			netmask: "255.255.255.0",
			cidr:    "",
		},
		"invalid netmask": {
			network: "192.168.0.0",
			netmask: "asdf",
			cidr:    "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			conf := New()
			conf.Network = tt.network
			conf.Netmask = tt.netmask
			assert.Equal(t, tt.cidr, conf.NetworkCIDR())
		})
	}
}
