package oci

import (
	"path/filepath"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

var defaultCachePath = filepath.Join(warewulfconf.Get().Warewulf.DataStore, "/container-cache/oci/")

const (
	blobPrefix   = "blobs"
	rootfsPrefix = "rootfs"
)
