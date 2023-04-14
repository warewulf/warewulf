package oci

import (
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"path/filepath"
)

var defaultCachePath = filepath.Join(warewulfconf.DataStore(), "/container-cache/oci/")

const (
	blobPrefix       = "blobs"
	rootfsPrefix     = "rootfs"
)
