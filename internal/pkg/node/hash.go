package node

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

/*
Calculate the hash of NodeYaml in an orderder fashion
*/
func (config *NodesYaml) Hash() [32]byte {
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

/*
Return the hash as string
*/
func (config *NodesYaml) StringHash() string {
	buffer := config.Hash()
	return hex.EncodeToString(buffer[:])
}
