package config

// A MountEntry represents a bind mount that is applied to an image
// during exec and shell.
type MountEntry struct {
	Source    string `yaml:"source"`
	Dest      string `yaml:"dest,omitempty"`
	ReadOnlyP *bool  `yaml:"readonly,omitempty"`
	Options   string `yaml:"options,omitempty"` // ignored at the moment
	CopyP     *bool  `yaml:"copy,omitempty"`    // temporarily copy the file into the image
}

func (mount MountEntry) ReadOnly() bool {
	return BoolP(mount.ReadOnlyP)
}

func (mount MountEntry) Copy() bool {
	return BoolP(mount.CopyP)
}
