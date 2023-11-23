package config

import (
	"github.com/creasty/defaults"
)

// NFSConf represents the NFS configuration that will be used by
// Warewulf to generate exports on the server and mounts on compute
// nodes.
type NFSConf struct {
	Enabled         bool             `yaml:"enabled" default:"true"`
	ExportsExtended []*NFSExportConf `yaml:"export paths" default:"[]"`
	SystemdName     string           `yaml:"systemd name" default:"nfsd"`
}

// An NFSExportConf reprents a single NFS export / mount.
type NFSExportConf struct {
	Path          string `yaml:"path" default:"/dev/null"`
	ExportOptions string `default:"rw,sync,no_subtree_check" yaml:"export options"`
	MountOptions  string `default:"defaults" yaml:"mount options"`
	Mount         bool   `default:"true" yaml:"mount"`
}

// Implements the Unmarshal interface for NFSConf to set default
// values.
func (conf *NFSConf) Unmarshal(unmarshal func(interface{}) error) error {
	if err := defaults.Set(conf); err != nil {
		return err
	}
	return nil
}
