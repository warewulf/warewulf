package imprt

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/types"
	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	source := args[0]

	// Shim in a name if none given.
	name := ""
	if len(args) == 2 {
		name = args[1]
	}
	if name == "" {
		name = path.Base(source)
		wwlog.Info("Setting image name: %s", name)
	}
	if !image.ValidName(name) {
		return fmt.Errorf("image name contains illegal characters: %s", name)
	}

	fullPath := image.SourceDir(name)

	// image already exists and should be removed first
	if util.IsDir(fullPath) {
		if SetUpdate {
			wwlog.Info("Updating existing image")
		} else if SetForce {
			wwlog.Info("Overwriting existing image")
			if err := os.RemoveAll(fullPath); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("image name exists, specify --force, --update, or choose a different name: %s", name)
		}
	}

	if strings.HasPrefix(source, "docker://") || strings.HasPrefix(source, "docker-daemon://") ||
		strings.HasPrefix(source, "file://") || util.IsFile(source) {
		var sCtx *types.SystemContext
		sCtx, err := image.GetSystemContext(OciNoHttps, OciUsername, OciPassword, Platform)
		if err != nil {
			return err
		}

		if util.IsFile(source) && !filepath.IsAbs(source) {
			source, err = filepath.Abs(source)
			if err != nil {
				return fmt.Errorf("when resolving absolute path of %s, err: %v", source, err)
			}
		}
		err = image.ImportDocker(source, name, sCtx)
		if err != nil {
			_ = image.DeleteSource(name)
			return fmt.Errorf("could not import image: %s", err.Error())
		}
	} else if util.IsDir(source) {
		if err := image.ImportDirectory(source, name); err != nil {
			_ = image.DeleteSource(name)
			return fmt.Errorf("could not import image: %s", err.Error())
		}
	} else {
		return fmt.Errorf("invalid dir or uri: %s", source)
	}

	if SyncUser {
		if err := image.Syncuser(name, true); err != nil {
			return fmt.Errorf("syncuser error: %w", err)
		}
	}

	if SetBuild {
		wwlog.Info("Building image: %s", name)
		if err := image.Build(name, true); err != nil {
			return fmt.Errorf("could not build image %s: %s", name, err.Error())
		}
	}
	return nil
}
