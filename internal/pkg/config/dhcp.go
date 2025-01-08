package config

// DHCPConf represents the configuration for the DHCP service that
// Warewulf will configure.
type DHCPConf struct {
	EnabledP    *bool  `yaml:"enabled,omitempty" default:"true"`
	Template    string `yaml:"template,omitempty" default:"default"`
	RangeStart  string `yaml:"range start,omitempty"`
	RangeEnd    string `yaml:"range end,omitempty"`
	SystemdName string `yaml:"systemd name,omitempty" default:"dhcpd"`
}

func (this DHCPConf) Enabled() bool {
	return BoolP(this.EnabledP)
}
