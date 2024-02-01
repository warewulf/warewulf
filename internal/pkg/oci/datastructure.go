package oci

import (
	"time"

	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
)

// Output is the output format of (skopeo inspect),
// primarily so that we can format it with a simple json.MarshalIndent.
type InspectOutput struct {
	Name          string `json:",omitempty"`
	Tag           string `json:",omitempty"`
	Digest        digest.Digest
	RepoTags      []string
	Created       *time.Time
	DockerVersion string
	Labels        map[string]string
	Architecture  string
	Os            string
	Layers        []string
	LayersData    []types.ImageInspectLayer
	Env           []string
}
