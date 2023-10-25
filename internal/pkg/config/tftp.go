package config

// TFTPConf represents that configuration for the TFTP service that
// Warewulf will configure.
type TFTPConf struct {
	Enabled     bool   `yaml:"enabled" default:"true"`
	TftpRoot    string `yaml:"tftproot" default:"/var/lib/tftpboot"`
	SystemdName string `yaml:"systemd name" default:"tftp"`

	// Path is relative to buildconfig.DATADIR()
	IpxeBinaries map[string]string `yaml:"ipxe" default:"{\"00:09\": \"ipxe/ipxe-snponly-x86_64.efi\",\"00:00\": \"ipxe/undionly.kpxe\",\"00:0B\": \"ipxe/arm64-efi/snponly.efi\",\"00:07\":  \"ipxe/ipxe-snponly-x86_64.efi\"}"`
}
