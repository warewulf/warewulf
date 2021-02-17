package imprt

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/containers/image/v5/types"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
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
		fmt.Printf("Setting VNFS name: %s\n", name)
	}

	if container.ValidName(name) == false {
		wwlog.Printf(wwlog.ERROR, "VNFS name contains illegal characters: %s\n", name)
		os.Exit(1)
	}

	fullPath := container.SourceDir(name)

	if util.IsDir(fullPath) == true {
		if SetForce == true {
			wwlog.Printf(wwlog.WARN, "Overwriting existing VNFS\n")
			err := os.RemoveAll(fullPath)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		} else if SetUpdate == true {
			wwlog.Printf(wwlog.WARN, "Updating existing VNFS\n")
		} else {
			wwlog.Printf(wwlog.ERROR, "VNFS Name exists, specify --force, --update, or choose a different name: %s\n", name)
			os.Exit(1)
		}
	}

	sCtx, err := getSystemContext()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
	}

	err = container.PullURI(uri, name, sCtx)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not pull image: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Building container: %s\n", name)
	output, err := container.Build(name, true)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not build container %s: %s\n", name, err)
		os.Exit(1)
	} else {
		fmt.Printf("%s: %s\n", name, output)
	}

	if SetDefault == true {
		nodeDB, err := node.New()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
			os.Exit(1)
		}

		//TODO: Don't loop through profiles, instead have a nodeDB function that goes directly to the map
		profiles, _ := nodeDB.FindAllProfiles()
		for _, profile := range profiles {
			wwlog.Printf(wwlog.DEBUG, "Looking for profile default: %s\n", profile.Id.Get())
			if profile.Id.Get() == "default" {
				wwlog.Printf(wwlog.DEBUG, "Found profile default, setting container name to: %s\n", name)
				profile.ContainerName.Set(name)
				nodeDB.ProfileUpdate(profile)
			}
		}
		nodeDB.Persist()
		fmt.Printf("Set default profile to container: %s\n", name)
		warewulfd.DaemonReload()

	}

	return nil
}
