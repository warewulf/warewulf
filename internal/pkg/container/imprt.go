package container

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/containers/image/v5/copy"

	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
	dircopy "github.com/containers/storage/drivers/copy"
	"github.com/pkg/errors"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/oci"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

/*
Import a container from the given URI as the given name. Also a
SystemContext has to be provided.
*/
func ImportDocker(uri string, name string, changes bool, sCtx *types.SystemContext, pCtx *signature.PolicyContext) error {
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
		oci.OptSetId(name),
		oci.OptSetRecordChanges(changes),
	)
	if err != nil {
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

	err = dircopy.DirCopy(uri, fullPath, dircopy.Content, true)
	if err != nil {
		return err
	}

	return nil
}

func ReimportContainer(src, name string, recordChanges bool, sCtx *types.SystemContext, pCtx *signature.PolicyContext) (err error) {
	OciBlobCacheDir := warewulfconf.Get().Warewulf.DataStore + "/oci"
	fullPath := SourceDir(name)
	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		return
	}
	p, err := oci.NewPuller(
		oci.OptSetBlobCachePath(OciBlobCacheDir),
		oci.OptSetSystemContext(sCtx),
		oci.OptSetPolicyContext(pCtx),
		oci.OptSetId(name),
	)
	if err != nil {
		return err
	}
	cacheRef, err := layout.ParseReference(OciBlobCacheDir + ":" + src)
	if err != nil {
		return fmt.Errorf("unable to generate local oci reference: %v", err)
	}
	dstRef, err := layout.ParseReference(OciBlobCacheDir + ":" + name)
	if err != nil {
		return fmt.Errorf("unable to generate local oci reference: %v", err)
	}
	_, err = copy.Image(context.Background(), pCtx, dstRef, cacheRef, &copy.Options{
		ReportWriter:     os.Stdout,
		SourceCtx:        sCtx,
		RemoveSignatures: false,
	})
	if err != nil {
		return err
	}
	if recordChanges {
		recRef, err := layout.ParseReference(OciBlobCacheDir + ":" + name + oci.CacheContainerSuffix)
		if err != nil {
			return fmt.Errorf("unable to generate local oci reference: %v", err)
		}
		_, err = copy.Image(context.Background(), pCtx, recRef, cacheRef, &copy.Options{
			ReportWriter:     os.Stdout,
			SourceCtx:        sCtx,
			RemoveSignatures: false,
		})
		if err != nil {
			return err
		}
	}
	return p.PullFromCache(context.Background(), cacheRef, fullPath)
}
