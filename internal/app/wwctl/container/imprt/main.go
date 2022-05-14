package imprt

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/containers/image/v5/types"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

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

func CobraRunE(cmd *cobra.Command, args []string) error {
	var name string
	uri := args[0]

	if len(args) == 2 {
		name = args[1]
	} else {
		name = path.Base(uri)
		wwlog.Info("Setting VNFS name: %s", name)
	}

	if !container.ValidName(name) {
		wwlog.Error("VNFS name contains illegal characters: %s", name)
		os.Exit(1)
	}

	fullPath := container.SourceDir(name)

	if util.IsDir(fullPath) {
		if SetForce {
			wwlog.Info("Overwriting existing VNFS")
			err := os.RemoveAll(fullPath)
			if err != nil {
				wwlog.ErrorExc(err, "")
				os.Exit(1)
			}
		} else if SetUpdate {
			wwlog.Info("Updating existing VNFS")
		} else {
			wwlog.Error("VNFS Name exists, specify --force, --update, or choose a different name: %s", name)
			os.Exit(1)
		}
	} else if strings.HasPrefix(uri, "docker://") || strings.HasPrefix(uri, "docker-daemon://") ||
		strings.HasPrefix(uri, "file://") || util.IsFile(uri) {
		sCtx, err := getSystemContext()
		if err != nil {
			wwlog.ErrorExc(err, "")
		}

		err = container.ImportDocker(uri, name, sCtx)
		if err != nil {
			wwlog.Error("Could not import image: %s", err)
			_ = container.DeleteSource(name)
			os.Exit(1)
		}
	} else if util.IsDir(uri) {
		err := container.ImportDirectory(uri, name)
		if err != nil {
			wwlog.Error("Could not import image: %s", err)
			_ = container.DeleteSource(name)
			os.Exit(1)
		}
	} else {
		wwlog.Error("Invalid dir or uri: %s", uri)
		os.Exit(1)
	}

	wwlog.Info("Updating the container's /etc/resolv.conf")
	err := util.CopyFile("/etc/resolv.conf", path.Join(container.RootFsDir(name), "/etc/resolv.conf"))
	if err != nil {
		wwlog.Warn("Could not copy /etc/resolv.conf into container: %s", err)
	}

	err = container.SyncUids(name, !SyncUser)
	if err != nil && !SyncUser {
		wwlog.Error("Error in user sync, fix error and run 'syncuser' manually: %s", err)
		os.Exit(1)
	}

	wwlog.Info("Building container: %s", name)
	err = container.Build(name, true)
	if err != nil {
		wwlog.Error("Could not build container %s: %s", name, err)
		os.Exit(1)
	}

	if SetDefault {
		nodeDB, err := node.New()
		if err != nil {
			wwlog.Error("Could not open node configuration: %s", err)
			os.Exit(1)
		}

		//TODO: Don't loop through profiles, instead have a nodeDB function that goes directly to the map
		profiles, _ := nodeDB.FindAllProfiles()
		for _, profile := range profiles {
			wwlog.Printf(wwlog.DEBUG, "Looking for profile default: %s", profile.Id.Get())
			if profile.Id.Get() == "default" {
				wwlog.Printf(wwlog.DEBUG, "Found profile default, setting container name to: %s", name)
				profile.ContainerName.Set(name)
				err := nodeDB.ProfileUpdate(profile)
				if err != nil {
					return errors.Wrap(err, "failed to update profile")
				}
			}
		}
		err = nodeDB.Persist()
		if err != nil {
			return errors.Wrap(err, "failed to persist nodedb")
		}

		wwlog.Info("Set default profile to container: %s", name)
		err = warewulfd.DaemonReload()
		if err != nil {
			return errors.Wrap(err, "failed to reload warewulf daemon")
		}
	}

	return nil
}
