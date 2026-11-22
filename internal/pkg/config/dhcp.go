package config

import (
	"github.com/warewulf/warewulf/internal/pkg/util"
)

// DHCPConf represents the configuration for the DHCP service that
// Warewulf will configure.
type DHCPConf struct {
	EnabledP *bool  `yaml:"enabled,omitempty" default:"true"`
	Template string `yaml:"template,omitempty" default:"default"`
	// RouterAdvertisements selects where IPv6 router advertisements
	// come from on the provisioning network. Only the dnsmasq backend
	// sends router advertisements; one of "warewulf" (the default: when
	// unset or unknown, dnsmasq advertises and acts as the default
	// router), "warewulf-no-default-route" (dnsmasq advertises with
	// router lifetime 0, so clients do not use the Warewulf server as
	// their default gateway), or "upstream" (dnsmasq sends no router
	// advertisements). Any other value is treated as "warewulf".
	RouterAdvertisements string `yaml:"router advertisements,omitempty"`
	RangeStart           string `yaml:"range start,omitempty"`
	RangeEnd             string `yaml:"range end,omitempty"`
	Range6Start          string `yaml:"range6 start,omitempty"`
	Range6End            string `yaml:"range6 end,omitempty"`
	SystemdName          string `yaml:"systemd name,omitempty" default:"dhcpd"`
	// Overlays are the host overlays applied when DHCP is configured.
	Overlays OverlayList `yaml:"overlays,omitempty"`
}

func (conf DHCPConf) Enabled() bool {
	return util.BoolP(conf.EnabledP)
}
