package image

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/containers/image/v5/types"
	"github.com/containers/storage/drivers/copy"
	"github.com/pkg/errors"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/oci"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func ImportDocker(uri string, name string, sCtx *types.SystemContext) error {
	OciBlobCacheDir := warewulfconf.Get().Paths.OciBlobCachedir()

	err := os.MkdirAll(OciBlobCacheDir, 0755)
	if err != nil {
		return err
	}

	if !ValidName(name) {
		return errors.New("Image name contains illegal characters: " + name)
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

// create the system context and reading out environment variables
func GetSystemContext(noHttps bool, username string, password string, platform string) (sCtx *types.SystemContext, err error) {
	sCtx = &types.SystemContext{}
	// only check env if noHttps wasn't set
	if !noHttps {
		val, ok := os.LookupEnv("WAREWULF_OCI_NOHTTPS")
		if ok {

			noHttps, err = strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf("while parsing insecure http option: %v", err)
			}

		}
		// only set this if we want to disable, otherwise leave as undefined
		if noHttps {
			sCtx.DockerInsecureSkipTLSVerify = types.NewOptionalBool(true)
		}
		sCtx.OCIInsecureSkipTLSVerify = noHttps
	}
	if username == "" {
		username, _ = os.LookupEnv("WAREWULF_OCI_USERNAME")
	}
	if password == "" {
		password, _ = os.LookupEnv("WAREWULF_OCI_PASSWORD")
	}
	if username != "" || password != "" {
		if username != "" && password != "" {
			sCtx.DockerAuthConfig = &types.DockerAuthConfig{
				Username: username,
				Password: password,
			}
		} else {
			return nil, fmt.Errorf("oci username and password env vars must be specified together")
		}
	}
	if platform == "" {
		platform, _ = os.LookupEnv("WAREWULF_OCI_PLATFORM")
	}
	if platform != "" {
		sCtx.ArchitectureChoice = platform
	}
	return sCtx, nil
}
