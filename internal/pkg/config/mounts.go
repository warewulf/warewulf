package config

// A MountEntry represents a bind mount that is applied to a container
// during exec and shell.
type MountEntry struct {
	Source    string `yaml:"source"`
	Dest      string `yaml:"dest,omitempty"`
	ReadOnlyP *bool  `yaml:"readonly,omitempty"`
	Options   string `yaml:"options,omitempty"` // ignored at the moment
	CopyP     *bool  `yaml:"copy,omitempty"`    // temporarily copy the file into the container
}

func (this MountEntry) ReadOnly() bool {
	return BoolP(this.ReadOnlyP)
}

func (this MountEntry) Copy() bool {
	return BoolP(this.CopyP)
}
