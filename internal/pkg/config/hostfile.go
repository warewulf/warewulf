package config

// HostfileConf represents the configuration for the host file
// (/etc/hosts) that Warewulf manages on the server.
type HostfileConf struct {
	// Overlays are the host overlays applied when the hostfile is
	// configured.
	Overlays OverlayList `yaml:"overlays,omitempty"`
}
