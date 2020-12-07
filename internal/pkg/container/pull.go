package container

import (
	"context"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/oci"
	"os"
)

func PullURI(uri string, name string) error {
	OciBlobCacheDir := config.LocalStateDir + "/oci/blobs"

	err := os.MkdirAll(OciBlobCacheDir, 0755)
	if err != nil {
		return err
	}

	if ValidName(name) == false {
		return errors.New("VNFS name contains illegal characters: " + name)
	}

	fullPath := RootFsDir(name)

	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return err
	}

	p, err := oci.NewPuller(
		oci.OptSetBlobCachePath(OciBlobCacheDir),
		oci.OptSetSystemContext(nil),
	)
	if err != nil {
		return err
	}

	p.GenerateID(context.Background(), uri)

	if err := p.Pull(context.Background(), uri, fullPath); err != nil {
		return err
	}

	return nil
}
