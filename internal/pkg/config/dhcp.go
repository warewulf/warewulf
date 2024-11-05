package config

// DHCPConf represents the configuration for the DHCP service that
// Warewulf will configure.
type DHCPConf struct {
	Enabled     bool   `yaml:"enabled" default:"true"`
	Template    string `yaml:"template,omitempty" default:"default"`
	RangeStart  string `yaml:"range start,omitempty"`
	RangeEnd    string `yaml:"range end,omitempty"`
	SystemdName string `yaml:"systemd name,omitempty" default:"dhcpd"`
}
