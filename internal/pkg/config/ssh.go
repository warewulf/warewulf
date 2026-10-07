package config

type SSHConf struct {
	KeyTypes []string `yaml:"key types,omitempty" default:"[\"ed25519\",\"ecdsa\",\"rsa\"]"`
	// Overlays are the host overlays applied when SSH is configured.
	Overlays OverlayList `yaml:"overlays,omitempty"`
}
