package oci

import (
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"path/filepath"
)

var defaultCachePath = filepath.Join(warewulfconf.Get().Warewulf.DataStore, "/container-cache/oci/")

const (
	blobPrefix   = "blobs"
	rootfsPrefix = "rootfs"
)
