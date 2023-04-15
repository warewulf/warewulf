package config


// TFTPConf represents that configuration for the TFTP service that
// Warewulf will configure.
type TFTPConf struct {
	Enabled      bool              `yaml:"enabled" default:"true"`
	TFTPRoot     string            `yaml:"tftproot" default:"/var/lib/tftpboot"`
	SystemdName  string            `yaml:"systemd name" default:"tftp"`

	// Path is relative to buildconfig.DATADIR()
	IPXEBinaries map[string]string `yaml:"ipxe" default:"{\"00:09\": \"x86_64.efi\",\"00:00\": \"x86_64.kpxe\",\"00:0B\": \"arm64.efi\",\"00:07\":  \"x86_64.efi\"}"`
}
