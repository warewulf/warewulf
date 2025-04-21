package image

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containers/image/v5/types"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func ImageBuild(cbp *wwapiv1.ImageBuildParameter) (err error) {
	if cbp == nil {
		return fmt.Errorf("input parameter is nil")
	}

	var images []string

	if cbp.All {
		images, err = image.ListSources()
	} else {
		images = cbp.ImageNames
	}

	if len(images) == 0 {
		return
	}

	for _, c := range images {
		if !image.ValidSource(c) {
			return fmt.Errorf("image name does not exist: %s", c)
		}

		err = image.Build(c, cbp.Force)
		if err != nil {
			return fmt.Errorf("could not build image %s: %s", c, err)
		}
	}
	return
}

func ImageDelete(cdp *wwapiv1.ImageDeleteParameter) (err error) {
	if cdp == nil {
		return fmt.Errorf("input parameter is nil")
	}

	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open nodeDB: %s", err)
	}

ARG_LOOP:
	for i := 0; i < len(cdp.ImageNames); i++ {
		//_, arg := range args {
		imageName := cdp.ImageNames[i]
		for _, n := range nodeDB.Nodes {
			if n.ImageName == imageName {
				wwlog.Error("image %s is in use by node %s, skipping", imageName, n.Id())
				continue ARG_LOOP
			}
		}
		for _, p := range nodeDB.NodeProfiles {
			if p.ImageName == imageName {
				wwlog.Error("image %s is in use by profile %s, skipping", imageName, p.Id())
				continue ARG_LOOP
			}
		}

		if !image.ValidSource(imageName) {
			wwlog.Error("image name is not a valid source: %s", imageName)
			continue
		}
		err := image.DeleteSource(imageName)
		if err != nil {
			wwlog.Error("could not remove source: %s", imageName)
		}
		err = image.DeleteImage(imageName)
		if err != nil {
			wwlog.Error("could not remove image files %s", imageName)
		}

		fmt.Printf("Image has been deleted: %s\n", imageName)
	}

	return
}

func ImageImport(cip *wwapiv1.ImageImportParameter) (imageName string, err error) {
	if cip == nil {
		err = fmt.Errorf("input parameter is nil")
		return
	}

	if cip.Name == "" {
		name := path.Base(cip.Source)
		wwlog.Info("Setting image name: %s", name)
		cip.Name = name
	}
	if !image.ValidName(cip.Name) {
		err = fmt.Errorf("image name contains illegal characters: %s", cip.Name)
		return
	}

	imageName = cip.Name
	fullPath := image.SourceDir(cip.Name)

	// image already exists and should be removed first
	if util.IsDir(fullPath) && cip.Force {
		wwlog.Info("Overwriting existing image")
		err = os.RemoveAll(fullPath)
		if err != nil {
			return
		}
	}

	if util.IsDir(fullPath) {
		if !cip.Update {
			err = fmt.Errorf("image name exists, specify --force, --update, or choose a different name: %s", cip.Name)
			return
		}
		wwlog.Info("Updating existing image")
	} else if strings.HasPrefix(cip.Source, "docker://") || strings.HasPrefix(cip.Source, "docker-daemon://") ||
		strings.HasPrefix(cip.Source, "file://") || util.IsFile(cip.Source) {
		var sCtx *types.SystemContext
		sCtx, err = GetSystemContext(cip.OciNoHttps, cip.OciUsername, cip.OciPassword, cip.Platform)
		if err != nil {
			return
		}

		if util.IsFile(cip.Source) && !filepath.IsAbs(cip.Source) {
			cip.Source, err = filepath.Abs(cip.Source)
			if err != nil {
				err = fmt.Errorf("when resolving absolute path of %s, err: %v", cip.Source, err)
				return
			}
		}
		err = image.ImportDocker(cip.Source, cip.Name, sCtx)
		if err != nil {
			err = fmt.Errorf("could not import image: %s", err.Error())
			_ = image.DeleteSource(cip.Name)
			return
		}
	} else if util.IsDir(cip.Source) {
		err = image.ImportDirectory(cip.Source, cip.Name)
		if err != nil {
			err = fmt.Errorf("could not import image: %s", err.Error())
			_ = image.DeleteSource(cip.Name)
			return
		}
	} else {
		err = fmt.Errorf("invalid dir or uri: %s", cip.Source)
		return
	}

	if cip.SyncUser {
		err = image.Syncuser(cip.Name, true)
		if err != nil {
			err = fmt.Errorf("syncuser error: %w", err)
			return
		}
	}

	if cip.Build {
		wwlog.Info("Building image: %s", cip.Name)
		err = image.Build(cip.Name, true)
		if err != nil {
			err = fmt.Errorf("could not build image %s: %s", cip.Name, err.Error())
			return
		}
	}
	return
}

func ImageList() (imageInfo []*wwapiv1.ImageInfo, err error) {
	var sources []string

	sources, err = image.ListSources()
	if err != nil {
		wwlog.Error("%s", err)
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("%s", err)
		return
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("%s", err)
		return
	}

	nodemap := make(map[string]int)
	for _, n := range nodes {
		nodemap[n.ImageName]++
	}

	for _, source := range sources {
		if nodemap[source] == 0 {
			nodemap[source] = 0
		}

		wwlog.Debug("Finding kernel version for: %s", source)
		kernel := kernel.FindKernels(source).Default()
		kernelVersion := ""
		if kernel != nil {
			kernelVersion = kernel.Version()
		}
		var creationTime uint64
		sourceStat, err := os.Stat(image.SourceDir(source))
		wwlog.Debug("Checking creation time for: %s,%v", image.SourceDir(source), sourceStat.ModTime())
		if err != nil {
			wwlog.Error("%s", err)
		} else {
			creationTime = uint64(sourceStat.ModTime().Unix())
		}
		var modTime uint64
		imageStat, err := os.Stat(image.ImageFile(source))
		if err == nil {
			modTime = uint64(imageStat.ModTime().Unix())
		}
		imgSize := 0
		if imgF, err := os.Stat(image.ImageFile(source)); err == nil {
			imgSize = int(imgF.Size())
		}
		imgCSize := 0
		if imgFC, err := os.Stat(image.ImageFile(source) + ".gz"); err == nil {
			imgCSize = int(imgFC.Size())
		}
		imageInfo = append(imageInfo, &wwapiv1.ImageInfo{
			Name:          source,
			NodeCount:     uint32(nodemap[source]),
			KernelVersion: kernelVersion,
			CreateDate:    creationTime,
			ModDate:       modTime,
			ImgSize:       uint64(imgSize),
			ImgSizeComp:   uint64(imgCSize),
		})

	}
	return
}

func ImageShow(csp *wwapiv1.ImageShowParameter) (response *wwapiv1.ImageShowResponse, err error) {
	imageName := csp.ImageName

	if !image.ValidName(imageName) {
		err = fmt.Errorf("%s is not a valid image name", imageName)
		return
	}

	rootFsDir := image.RootFsDir(imageName)
	if !util.IsDir(rootFsDir) {
		err = fmt.Errorf("%s is not a valid image", imageName)
		return
	}
	kernel := kernel.FindKernels(imageName).Default()
	kernelVersion := ""
	if kernel != nil {
		kernelVersion = kernel.Version()
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
		if n.ImageName == imageName {
			nodeList = append(nodeList, n.Id())
		}
	}

	response = &wwapiv1.ImageShowResponse{
		Name:          imageName,
		Rootfs:        rootFsDir,
		Nodes:         nodeList,
		KernelVersion: kernelVersion,
	}
	return
}

func ImageRename(crp *wwapiv1.ImageRenameParameter) (err error) {
	// rename the image source folder
	sourceDir := image.SourceDir(crp.ImageName)
	destDir := image.SourceDir(crp.TargetName)
	err = os.Rename(sourceDir, destDir)
	if err != nil {
		return err
	}

	err = image.DeleteImage(crp.ImageName)
	if err != nil {
		wwlog.Warn("Could not remove image files for %s: %s", crp.ImageName, err)
	}

	if crp.Build {
		err = image.Build(crp.TargetName, true)
		if err != nil {
			return err
		}
	}

	// update the nodes profiles image name
	nodeDB, err := node.New()
	if err != nil {
		return err
	}

	for nodeId, node := range nodeDB.Nodes {
		if node.ImageName == crp.ImageName {
			wwlog.Debug("updating node %s image to %s", nodeId, crp.TargetName)
			nodeDB.Nodes[nodeId].ImageName = crp.TargetName
		}
	}

	for profileId, profile := range nodeDB.NodeProfiles {
		if profile.ImageName == crp.ImageName {
			wwlog.Debug("updating profile %s image to %s", profileId, crp.TargetName)
			nodeDB.NodeProfiles[profileId].ImageName = crp.TargetName
		}
	}

	err = nodeDB.Persist()
	if err != nil {
		return err
	}

	err = warewulfd.DaemonStatus()
	if err != nil {
		// warewulfd is not running, skip
		return nil
	}

	// else reload daemon to apply new changes
	return warewulfd.DaemonReload()
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
