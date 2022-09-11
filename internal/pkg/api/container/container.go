package container

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/containers/image/v5/types"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

func ContainerBuild(cbp *wwapiv1.ContainerBuildParameter) (err error) {

	if cbp == nil {
		return fmt.Errorf("ContainerBuildParameter is nil")
	}

	var containers []string

	if cbp.All {
		containers, err = container.ListSources()
	} else {
		containers = cbp.ContainerNames
	}

	if len(containers) == 0 {
		return
	}

	for _, c := range containers {
		if !container.ValidSource(c) {
			err = fmt.Errorf("VNFS name does not exist: %s", c)
			wwlog.Error("%s\n", err)
			return
		}

		err = container.Build(c, cbp.Force)
		if err != nil {
			wwlog.Error("Could not build container %s: %s\n", c, err)
			return
		}
	}

	if cbp.Default {
		if len(containers) != 1 {
			wwlog.Error("Can only set default for one container\n")
		} else {
			var nodeDB node.NodeYaml
			nodeDB, err = node.New()
			if err != nil {
				wwlog.Error("Could not open node configuration: %s\n", err)
				return
			}

			//TODO: Don't loop through profiles, instead have a nodeDB function that goes directly to the map
			profiles, _ := nodeDB.FindAllProfiles()
			for _, profile := range profiles {
				wwlog.Debug("Looking for profile default: %s\n", profile.Id.Get())
				if profile.Id.Get() == "default" {
					wwlog.Debug("Found profile default, setting container name to: %s\n", containers[0])
					profile.ContainerName.Set(containers[0])
					err := nodeDB.ProfileUpdate(profile)
					if err != nil {
						return errors.Wrap(err, "failed to update node profile")
					}
				}
			}
			// TODO: Need a wrapper and flock around this. Sometimes we restart warewulfd and sometimes we don't.
			err = nodeDB.Persist()
			if err != nil {
				return errors.Wrap(err, "failed to persist nodedb")
			}
			fmt.Printf("Set default profile to container: %s\n", containers[0])
		}
	}
	return
}

func ContainerDelete(cdp *wwapiv1.ContainerDeleteParameter) (err error) {

	if cdp == nil {
		return fmt.Errorf("ContainerDeleteParameter is nil")
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open nodeDB: %s\n", err)
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return
	}

ARG_LOOP:
	for i := 0; i < len(cdp.ContainerNames); i++ {
		//_, arg := range args {
		containerName := cdp.ContainerNames[i]
		for _, n := range nodes {
			if n.ContainerName.Get() == containerName {
				wwlog.Error("Container is configured for nodes, skipping: %s\n", containerName)
				continue ARG_LOOP
			}
		}

		if !container.ValidSource(containerName) {
			wwlog.Error("Container name is not a valid source: %s\n", containerName)
			continue
		}
		err := container.DeleteSource(containerName)
		if err != nil {
			wwlog.Error("Could not remove source: %s\n", containerName)
		} else {
			fmt.Printf("Container has been deleted: %s\n", containerName)
		}
	}

	return
}

func ContainerImport(cip *wwapiv1.ContainerImportParameter) (containerName string, err error) {

	if cip == nil {
		err = fmt.Errorf("NodeAddParameter is nil")
		return
	}

	if cip.Name == "" {
		name := path.Base(cip.Source)
		wwlog.Info("Setting VNFS name: %s", name)
		cip.Name = name
	}
	if !container.ValidName(cip.Name) {
		err = fmt.Errorf("VNFS name contains illegal characters: %s", cip.Name)
		wwlog.Error(err.Error())
		return
	}

	containerName = cip.Name
	fullPath := container.SourceDir(cip.Name)

	if util.IsDir(fullPath) {
		if cip.Force {
			wwlog.Info("Overwriting existing VNFS")
			err = os.RemoveAll(fullPath)
			if err != nil {
				wwlog.ErrorExc(err, "")
				return
			}
		} else if cip.Update {
			wwlog.Info("Updating existing VNFS")
		} else {
			err = fmt.Errorf("VNFS Name exists, specify --force, --update, or choose a different name: %s", cip.Name)
			wwlog.Error(err.Error())
			return
		}
	} else if strings.HasPrefix(cip.Source, "docker://") || strings.HasPrefix(cip.Source, "docker-daemon://") ||
		strings.HasPrefix(cip.Source, "file://") || util.IsFile(cip.Source) {
		var sCtx *types.SystemContext
		sCtx, err = getSystemContext()
		if err != nil {
			wwlog.ErrorExc(err, "")
			// TODO: mhink - return was missing here. Was that deliberate?
		}

		err = container.ImportDocker(cip.Source, cip.Name, sCtx)
		if err != nil {
			err = fmt.Errorf("could not import image: %s", err.Error())
			wwlog.Error(err.Error())
			_ = container.DeleteSource(cip.Name)
			return
		}
	} else if util.IsDir(cip.Source) {
		err = container.ImportDirectory(cip.Source, cip.Name)
		if err != nil {
			err = fmt.Errorf("could not import image: %s", err.Error())
			wwlog.Error(err.Error())
			_ = container.DeleteSource(cip.Name)
			return
		}
	} else {
		err = fmt.Errorf("invalid dir or uri: %s", cip.Source)
		wwlog.Error(err.Error())
		return
	}

	wwlog.Info("Updating the container's /etc/resolv.conf")
	err = util.CopyFile("/etc/resolv.conf", path.Join(container.RootFsDir(cip.Name), "/etc/resolv.conf"))
	if err != nil {
		wwlog.Warn("Could not copy /etc/resolv.conf into container: %s\n", err)
	}

	err = container.SyncUids(cip.Name, !cip.SyncUser)
	if err != nil && !cip.SyncUser {
		err = fmt.Errorf("error in user sync, fix error and run 'syncuser' manually: %s", err)
		wwlog.Error(err.Error())
		return
	}

	wwlog.Info("Building container: %s", cip.Name)
	err = container.Build(cip.Name, true)
	if err != nil {
		err = fmt.Errorf("could not build container %s: %s", cip.Name, err.Error())
		wwlog.Error(err.Error())
		return
	}

	if cip.Default {
		var nodeDB node.NodeYaml
		nodeDB, err = node.New()
		if err != nil {
			err = fmt.Errorf("could not open node configuration: %s", err.Error())
			wwlog.Error(err.Error())
			return
		}

		//TODO: Don't loop through profiles, instead have a nodeDB function that goes directly to the map
		profiles, _ := nodeDB.FindAllProfiles()
		for _, profile := range profiles {
			wwlog.Debug("Looking for profile default: %s", profile.Id.Get())
			if profile.Id.Get() == "default" {
				wwlog.Debug("Found profile default, setting container name to: %s", cip.Name)
				profile.ContainerName.Set(cip.Name)
				err = nodeDB.ProfileUpdate(profile)
				if err != nil {
					err = errors.Wrap(err, "failed to update profile")
					return
				}
			}
		}
		// TODO: We need this in a function with a flock around it.
		// Also need to understand if the daemon restart is only to
		// reload the config or if there is something more.
		err = nodeDB.Persist()
		if err != nil {
			err = errors.Wrap(err, "failed to persist nodedb")
			return
		}

		wwlog.Info("Set default profile to container: %s", cip.Name)
		err = warewulfd.DaemonReload()
		if err != nil {
			err = errors.Wrap(err, "failed to reload warewulf daemon")
			return
		}
	}
	return
}

func ContainerList() (containerInfo []*wwapiv1.ContainerInfo, err error) {
	var sources []string

	sources, err = container.ListSources()
	if err != nil {
		wwlog.Error("%s\n", err)
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("%s\n", err)
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("%s\n", err)
		return
	}

	nodemap := make(map[string]int)
	for _, n := range nodes {
		nodemap[n.ContainerName.Get()]++
	}

	for _, source := range sources {
		if nodemap[source] == 0 {
			nodemap[source] = 0
		}

		wwlog.Debug("Finding kernel version for: %s\n", source)
		kernelVersion := container.KernelVersion(source)

		containerInfo = append(containerInfo, &wwapiv1.ContainerInfo{
			Name:          source,
			NodeCount:     uint32(nodemap[source]),
			KernelVersion: kernelVersion,
		})
	}
	return
}

func ContainerShow(csp *wwapiv1.ContainerShowParameter) (response *wwapiv1.ContainerShowResponse, err error) {

	containerName := csp.ContainerName

	if !container.ValidName(containerName) {
		err = fmt.Errorf("%s is not a valid container", containerName)
		return
	}

	rootFsDir := container.RootFsDir(containerName)

	kernelVersion := container.KernelVersion(containerName)
	if kernelVersion != "" {
		kernelVersion = "not found"
		fmt.Printf("Kernelversion: %s\n", kernelVersion)
	}

	nodeDB, err := node.New()
	if err != nil {
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		return
	}

	var nodeList []string
	for _, n := range nodes {
		if n.ContainerName.Get() == containerName {

			nodeList = append(nodeList, n.Id.Get())
		}
	}

	response = &wwapiv1.ContainerShowResponse{
		Name:          containerName,
		Rootfs:        rootFsDir,
		Nodes:         nodeList,
		KernelVersion: kernelVersion,
	}
	return
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
