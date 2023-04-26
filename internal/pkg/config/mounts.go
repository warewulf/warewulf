package config

// A MountEntry represents a bind mount that is applied to a container
// during exec and shell.
type MountEntry struct {
	Source   string `yaml:"source" default:"/etc/resolv.conf"`
	Dest     string `yaml:"dest,omitempty" default:"/etc/resolv.conf"`
	ReadOnly bool   `yaml:"readonly,omitempty" default:"false"`
	Options  string `yaml:"options,omitempty"` // ignored at the moment
}
