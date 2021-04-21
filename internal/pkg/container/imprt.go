package container

import (
	"context"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"os"
	"path"

	"github.com/containers/image/v5/types"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/oci"
	"github.com/docker/docker/daemon/graphdriver/copy"
)

func ImportDocker(uri string, name string, sCtx *types.SystemContext) error {
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
		oci.OptSetSystemContext(sCtx),
	)
	if err != nil {
		return err
	}

	if _, err := p.GenerateID(context.Background(), uri); err != nil {
		return err
	}

	if err := p.Pull(context.Background(), uri, fullPath); err != nil {
		return err
	}

	return nil
}

func ImportDirectory(uri string, name string) error {
	fullPath := RootFsDir(name)

	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		return err
	}

	if !util.IsDir(uri) {
		return errors.New("Import directory does not exist: " + uri)
	}

	if !util.IsFile(path.Join(uri, "/bin/sh")) {
		return errors.New("Source directory has no /bin/sh: " + uri)
	}

	err = copy.DirCopy(uri, fullPath, copy.Content, true)
	if err != nil {
		return err
	}

	return nil
}
