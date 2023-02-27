package node

import (
	"crypto/sha256"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

/*
Calculate the hash of NodeYaml in an orderder fashion
*/
func (config *NodeYaml) Hash() [32]byte {
	// flatten out profiles and nodes
	for _, val := range config.NodeProfiles {
		val.Flatten()
	}
	for _, val := range config.Nodes {
		val.Flatten()
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		wwlog.Warn("couldn't marshall NodeYaml for hashing")
	}
	return sha256.Sum256(data)
}
