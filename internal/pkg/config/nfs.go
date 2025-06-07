package config

import (
	"github.com/creasty/defaults"
)

// NFSConf represents the NFS configuration that will be used by
// Warewulf to generate exports on the server and mounts on compute
// nodes.
type NFSConf struct {
	EnabledP               *bool                `yaml:"enabled,omitempty" default:"true"`
	ExportsExtended        []*NFSExportConf     `yaml:"export paths,omitempty" default:"[]"`
	GaneshaExportsExtended []*GaneshaExportConf `yaml:"ganesha exports,omitempty" default:"[]"`
	SystemdName            string               `yaml:"systemd name,omitempty" default:"nfsd"`
}

func (conf NFSConf) Enabled() bool {
	return BoolP(conf.EnabledP)
}

// An NFSExportConf reprents a single NFS export / mount.
type NFSExportConf struct {
	Path          string `yaml:"path" default:"/dev/null"`
	ExportOptions string `yaml:"export options,omitempty" default:"rw,sync,no_subtree_check"`
}

// A GaneshaExportConf represents a single NFS ganesha export
type GaneshaExportConf struct {
	Path       string `yaml:"path" default:"/dev/null"`
	Pseudo     string `yaml:"pseudo,omitempty"`
	AccessType string `yaml:"access type,omitempty" default:"rw"`
	Squash     string `yaml:"squash,omitempty" default:"root_squash"`
}

// Implements the Unmarshal interface for NFSConf to set default
// values.
func (conf *NFSConf) Unmarshal(unmarshal func(interface{}) error) error {
	if err := defaults.Set(conf); err != nil {
		return err
	}
	return nil
}
