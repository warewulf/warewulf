package container

import (
	"context"
	"os"
	"path"

	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage/drivers/copy"
	"github.com/pkg/errors"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/oci"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

/*
Import a container from the given URI as the given name. Also a
SystemContext has to be provided.
*/
func ImportDocker(uri string, name string, sCtx *types.SystemContext, pCtx *signature.PolicyContext) error {
	OciBlobCacheDir := warewulfconf.Get().Warewulf.DataStore + "/oci"

	err := os.MkdirAll(OciBlobCacheDir, 0755)
	if err != nil {
		return err
	}

	if !ValidName(name) {
		return errors.New("VNFS name contains illegal characters: " + name)
	}

	fullPath := SourceDir(name)

	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return err
	}
	p, err := oci.NewPuller(
		oci.OptSetBlobCachePath(OciBlobCacheDir),
		oci.OptSetSystemContext(sCtx),
		oci.OptSetPolicyContext(pCtx),
	)
	if err != nil {
		return err
	}
	/*
		if _, err := p.GenerateID(context.Background(), uri); err != nil {
			return err
		}
	*/
	p.SetId(name)
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

func ReimportContainer(inspectData oci.InspectOutput, name string, sCtx *types.SystemContext, pCtx *signature.PolicyContext) (err error) {
	OciBlobCacheDir := warewulfconf.Get().Warewulf.DataStore + "/oci"
	err = os.MkdirAll(OciBlobCacheDir, 0755)
	if err != nil {
		return err
	}
	fullPath := SourceDir(name)
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return
	}
	p, err := oci.NewPuller(
		oci.OptSetBlobCachePath(OciBlobCacheDir),
		oci.OptSetSystemContext(sCtx),
		oci.OptSetPolicyContext(pCtx),
	)
	if err != nil {
		return err
	}

	if err := p.PullFromCache(context.Background(), inspectData, fullPath); err != nil {
		return err
	}

	return nil
}
