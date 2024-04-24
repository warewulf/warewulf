package container

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containers/image/v5/types"
	"github.com/containers/storage/drivers/copy"
	"github.com/pkg/errors"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/oci"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type ImportParameter struct {
	Source                                  string
	Name                                    string
	Force, Update, Build, Default, SyncUser bool
}

func Import(param *ImportParameter) error {
	if param.Name == "" {
		if param.Source == "" {
			return fmt.Errorf("source is empty, unable to create name from it")
		}
		name := path.Base(param.Source)
		wwlog.Info("Setting VNFS name: %s", name)
		param.Name = name
	}

	if !ValidName(param.Name) {
		return fmt.Errorf("VNFS name contains illegal characters: %s", param.Name)
	}

	fullPath := SourceDir(param.Name)

	// container already exists and should be removed first
	if util.IsDir(fullPath) && param.Force {
		wwlog.Info("Overwriting existing VNFS")
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("while removing %s, failure with err: %s", fullPath, err)
		}
	}

	if util.IsDir(fullPath) {
		if !param.Update {
			return fmt.Errorf("VNFS Name exists, specify --force, --update, or choose a different name: %s", param.Name)
		}
		wwlog.Info("Updating existing VNFS")
	} else if strings.HasPrefix(param.Source, "docker://") || strings.HasPrefix(param.Source, "docker-daemon://") ||
		strings.HasPrefix(param.Source, "file://") || util.IsFile(param.Source) {
		var sCtx *types.SystemContext
		sCtx, err := getSystemContext()
		if err != nil {
			return fmt.Errorf("failed to retrieve system context, err: %s", err)
		}

		if util.IsFile(param.Source) && !filepath.IsAbs(param.Source) {
			param.Source, err = filepath.Abs(param.Source)
			if err != nil {
				return fmt.Errorf("when resolving absolute path of %s, err: %v", param.Source, err)
			}
		}

		err = ImportDocker(param.Source, param.Name, sCtx)
		if err != nil {
			defer func() {
				_ = DeleteSource(param.Name)
			}()
			return fmt.Errorf("could not import image: %s", err)
		}
	} else if util.IsDir(param.Source) {
		err := ImportDirectory(param.Source, param.Name)
		if err != nil {
			defer func() {
				_ = DeleteSource(param.Name)
			}()
			return fmt.Errorf("could not import image: %s", err)
		}
	} else {
		return fmt.Errorf("invalid dir or uri: %s", param.Source)
	}

	SyncUserShowOnly := !param.SyncUser
	err := SyncUids(param.Name, SyncUserShowOnly)
	if err != nil {
		err = fmt.Errorf("error in user sync, fix error and run 'syncuser' manually: %s", err)
		wwlog.Error(err.Error())
		if param.SyncUser {
			return err
		}
	}

	if param.Build {
		wwlog.Info("Building container: %s", param.Name)
		err = Build(&BuildParameter{
			Names: []string{param.Name},
			Force: true,
		})
		if err != nil {
			return fmt.Errorf("could not build container %s: %s", param.Name, err)
		}
	}

	if param.Default {
		wwlog.Info("Set default profile to container: %s", param.Name)
		err := SetProfileDefaultContainer(param.Name)
		if err != nil {
			return fmt.Errorf("failed to set default container to profile, err: %s", err)
		}
	}

	return nil
}

func ImportDocker(uri string, name string, sCtx *types.SystemContext) error {
	OciBlobCacheDir := warewulfconf.Get().Warewulf.DataStore + "/oci"

	err := os.MkdirAll(OciBlobCacheDir, 0755)
	if err != nil {
		return err
	}

	if !ValidName(name) {
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

// Private helpers

func setOCICredentials(sCtx *types.SystemContext) error {
	username, userSet := os.LookupEnv("WAREWULF_OCI_USERNAME")
	password, passSet := os.LookupEnv("WAREWULF_OCI_PASSWORD")
	if userSet || passSet {
		if userSet && passSet {
			sCtx.DockerAuthConfig = &types.DockerAuthConfig{
				Username: username,
				Password: password,
			}
		} else {
			return fmt.Errorf("oci username and password env vars must be specified together")
		}
	}
	return nil
}

func setNoHTTPSOpts(sCtx *types.SystemContext) error {
	val, ok := os.LookupEnv("WAREWULF_OCI_NOHTTPS")
	if !ok {
		return nil
	}

	noHTTPS, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("while parsing insecure http option: %v", err)
	}

	// only set this if we want to disable, otherwise leave as undefined
	if noHTTPS {
		sCtx.DockerInsecureSkipTLSVerify = types.NewOptionalBool(true)
	}
	sCtx.OCIInsecureSkipTLSVerify = noHTTPS

	return nil
}

func getSystemContext() (sCtx *types.SystemContext, err error) {
	sCtx = &types.SystemContext{}

	if err := setOCICredentials(sCtx); err != nil {
		return nil, err
	}

	if err := setNoHTTPSOpts(sCtx); err != nil {
		return nil, err
	}

	return sCtx, nil
}
