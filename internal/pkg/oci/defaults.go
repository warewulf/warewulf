package oci

import (
	"path/filepath"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

var defaultCachePath = filepath.Join(warewulfconf.Config("datastore"), "container-cache", "oci")

const (
	blobPrefix   = "blobs"
	rootfsPrefix = "rootfs"
)
