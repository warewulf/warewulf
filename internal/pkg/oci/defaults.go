package oci

import 	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
import  "path/filepath"

var defaultCachePath = filepath.Join(warewulfconf.DataStore(), "/container-cache/oci/")

const (
	blobPrefix       = "blobs"
	rootfsPrefix     = "rootfs"
)