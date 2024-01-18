package config

// WarewulfConf adds additional Warewulf-specific configuration to
// BaseConf.
type WarewulfConf struct {
	Port              int    `yaml:"port" default:"9983"`
	Secure            bool   `yaml:"secure" default:"true"`
	UpdateInterval    int    `yaml:"update interval" default:"60"`
	AutobuildOverlays bool   `yaml:"autobuild overlays" default:"true"`
	EnableHostOverlay bool   `yaml:"host overlay" default:"true"`
	Syslog            bool   `yaml:"syslog" default:"false"`
	DataStore         string `yaml:"datastore" default:"/var/lib/warewulf"`
	GrubBoot          bool   `yaml:"grubboot" default:"false"`
}
