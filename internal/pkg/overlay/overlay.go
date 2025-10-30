package overlay

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/coreos/go-systemd/v22/unit"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var ErrDoesNotExist = fmt.Errorf("overlay does not exist")

// Overlay represents an overlay directory path.
type Overlay string

// Name returns the base name of the overlay directory.
//
// This is derived from the full path of the overlay.
func (overlay Overlay) Name() string {
	return path.Base(overlay.Path())
}

// Path returns the string representation of the overlay path.
//
// This method allows the Overlay type to be easily converted back to its
// underlying string representation.
func (overlay Overlay) Path() string {
	return string(overlay)
}

// Rootfs returns the path to the root filesystem (rootfs) within the overlay.
//
// If the "rootfs" directory exists inside the overlay path, it returns the
// path to the "rootfs" directory. Otherwise, it checks if the overlay path
// itself is a directory and returns that. If neither exists, it defaults to
// returning the "rootfs" path.
func (overlay Overlay) Rootfs() string {
	rootfs := path.Join(overlay.Path(), "rootfs")
	if util.IsDir(rootfs) {
		return rootfs
	} else if util.IsDir(overlay.Path()) {
		return overlay.Path()
	} else {
		return rootfs
	}
}

// File constructs a full path to a file within the overlay's root filesystem.
//
// Parameters:
//   - filePath: The relative path of the file within the overlay.
//
// Returns:
//   - The full path to the specified file in the overlay's rootfs.
//     If the specified path is not contained within the overlay, the empty string is returned.
func (overlay Overlay) File(filePath string) string {
	rootfs := overlay.Rootfs()
	fullPath := path.Join(rootfs, filePath)
	cleanPath := filepath.Clean(fullPath)
	cleanRootfs := filepath.Clean(rootfs)
	rel, err := filepath.Rel(cleanRootfs, cleanPath)
	if err != nil {
		return ""
	}

	if strings.HasPrefix(rel, "..") {
		return ""
	}

	return cleanPath
}

// Exists checks whether the overlay path exists and is a directory.
//
// Returns:
//   - true if the overlay path exists and is a directory; false otherwise.
func (overlay Overlay) Exists() bool {
	return util.IsDir(overlay.Path())
}

// IsSiteOverlay determines whether the overlay is a site overlay.
//
// A site overlay is identified by its parent directory matching the configured
// site overlay directory path.
//
// Returns:
//   - true if the overlay is a site overlay; false otherwise.
func (overlay Overlay) IsSiteOverlay() bool {
	siteDir := filepath.Clean(config.Get().Paths.SiteOverlaydir())
	overlayPath := filepath.Clean(overlay.Path())
	if rel, err := filepath.Rel(siteDir, overlayPath); err != nil {
		return false
	} else {
		return !strings.HasPrefix(rel, "..")
	}
}

// IsDistributionOverlay determines whether the overlay is a distribution overlay.
//
// A distribution overlay is identified by its parent directory matching the configured
// distribution overlay directory path.
//
// Returns:
//   - true if the overlay is a distribution overlay; false otherwise.
func (overlay Overlay) IsDistributionOverlay() bool {
	siteDir := filepath.Clean(config.Get().Paths.DistributionOverlaydir())
	overlayPath := filepath.Clean(overlay.Path())
	if rel, err := filepath.Rel(siteDir, overlayPath); err != nil {
		return false
	} else {
		return !strings.HasPrefix(rel, "..")
	}
}

func (overlay Overlay) AddFile(filePath string, content []byte, parents bool, force bool) error {
	wwlog.Info("Creating file %s in overlay %s, force: %v", filePath, overlay.Name(), force)

	if !overlay.IsSiteOverlay() {
		siteOverlay, err := overlay.CloneToSite()
		if err != nil {
			return fmt.Errorf("failed to clone distribution overlay '%s' to site overlay: %w", overlay.Name(), err)
		}
		// replace the overlay with newly created siteOverlay
		overlay = siteOverlay
	}
	fullPath := overlay.File(filePath)
	// create necessary parent directories
	if parents {
		if err := os.MkdirAll(path.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("failed to create parent directories for %s: %w", fullPath, err)
		}
	}

	// if the file already exists and force is false, return an error
	if util.IsFile(fullPath) {
		if force {
			return os.WriteFile(fullPath, content, 0o644)
		}
		return fmt.Errorf("file %s already exists in overlay %s", filePath, overlay.Name())
	}

	return os.WriteFile(fullPath, content, 0o644)
}

func (overlay Overlay) Delete(force bool) (err error) {
	wwlog.Info("Deleting overlay %s, force: %v", overlay.Name(), force)
	if overlay.IsDistributionOverlay() {
		return fmt.Errorf("cannot delete a distribution overlay: %s", overlay.Name())
	}
	if force {
		err := os.RemoveAll(overlay.Path())
		if err != nil {
			return fmt.Errorf("failed to delete overlay forcely: %w", err)
		}
	} else {
		// remove rootfs at first
		if err = os.Remove(overlay.Rootfs()); err != nil {
			return fmt.Errorf("failed to delete overlay: %w", err)
		}
		if overlay.Exists() {
			if err = os.Remove(overlay.Path()); err != nil {
				return fmt.Errorf("failed to delete overlay: %w", err)
			}
		}
	}
	return nil
}

// DeleteFile deletes a file or the entire overlay directory.
// before deletion.
func (overlay Overlay) DeleteFile(filePath string, force, cleanup bool) (err error) {
	wwlog.Info("Deleting file %s from overlay %s, force: %v, cleanup: %v", filePath, overlay.Name(), force, cleanup)
	// first check if file exists
	if !util.IsFile(overlay.File(filePath)) {
		return fmt.Errorf("file %s does not exist in overlay %s", filePath, overlay.Name())
	}
	if overlay.IsDistributionOverlay() {
		siteOverlay, err := overlay.CloneToSite()
		if err != nil {
			return fmt.Errorf("failed to clone distribution overlay '%s' to site overlay: %w", overlay.Name(), err)
		}
		// replace the overlay with newly created siteOverlay
		overlay = siteOverlay
	}
	fullPath := overlay.File(filePath)
	if force {
		if err := os.RemoveAll(fullPath); err != nil {
			return fmt.Errorf("failed to delete file %s from overlay %s: %w", filePath, overlay.Name(), err)
		}
	} else {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to delete file %s from overlay %s: %w", filePath, overlay.Name(), err)
		}
	}

	if cleanup {
		// cleanup the empty parents
		i := path.Dir(fullPath)
		for i != overlay.Rootfs() {
			wwlog.Debug("Evaluating directory to remove: %s", i)
			err := os.Remove(i)
			if err != nil {
				// if the directory is not empty, we stop here
				if !os.IsNotExist(err) {
					wwlog.Debug("Could not remove directory %s: %v", i, err)
				}
				break
			}
			wwlog.Debug("Removed empty directory: %s", i)
			i = path.Dir(i)
		}
	}
	return nil
}

// chmod for the given path in the overlay
func (overlay Overlay) Chmod(path string, mode uint64) (err error) {
	if !overlay.IsSiteOverlay() {
		overlay, err = overlay.CloneToSite()
		if err != nil {
			return err
		}
	}
	fullPath := overlay.File(path)
	if !(util.IsFile(fullPath) || util.IsDir(fullPath)) {
		return fmt.Errorf("file does not exist within overlay: %s:%s", overlay.Name(), fullPath)
	}

	return os.Chmod(fullPath, os.FileMode(mode))
}

// chown file or dir in overlay
func (overlay Overlay) Chown(path string, uid, gid int) (err error) {
	if !overlay.IsSiteOverlay() {
		overlay, err = overlay.CloneToSite()
		if err != nil {
			return err
		}
	}
	fullPath := overlay.File(path)
	if !(util.IsFile(fullPath) || util.IsDir(fullPath)) {
		return fmt.Errorf("file does not exist within overlay: %s:%s", overlay.Name(), fullPath)
	}
	return os.Chown(fullPath, uid, gid)
}

func (overlay Overlay) Mkdir(path string, mode int32) (err error) {
	if !overlay.IsSiteOverlay() {
		overlay, err = overlay.CloneToSite()
		if err != nil {
			return err
		}
	}
	fullPath := overlay.File(path)
	if util.IsFile(fullPath) || util.IsDir(fullPath) {
		wwlog.Warn("path already exists, overwriting permissions: %s:%s", overlay.Name(), fullPath)
	}
	return os.MkdirAll(fullPath, os.FileMode(mode))
}

func BuildAllOverlays(nodes []node.Node, allNodes []node.Node, workerCount int) error {
	nodeChan := make(chan node.Node, len(nodes))
	errChan := make(chan error, len(nodes)*2)

	var wg sync.WaitGroup
	worker := func() {
		for n := range nodeChan {
			wwlog.Info("Building system overlay image for %s", n.Id())
			wwlog.Debug("System overlays for %s: [%s]", n.Id(), strings.Join(n.SystemOverlay, ", "))
			if len(n.SystemOverlay) < 1 {
				wwlog.Warn("No system overlays defined for %s", n.Id())
			}
			if err := BuildOverlay(n, allNodes, "system", n.SystemOverlay); err != nil {
				errChan <- fmt.Errorf("could not build system overlays %v for node %s: %w", n.SystemOverlay, n.Id(), err)
			}

			wwlog.Info("Building runtime overlay image for %s", n.Id())
			wwlog.Debug("Runtime overlays for %s: [%s]", n.Id(), strings.Join(n.RuntimeOverlay, ", "))
			if len(n.RuntimeOverlay) < 1 {
				wwlog.Warn("No runtime overlays defined for %s", n.Id())
			}
			if err := BuildOverlay(n, allNodes, "runtime", n.RuntimeOverlay); err != nil {
				errChan <- fmt.Errorf("could not build runtime overlays %v for node %s: %w", n.RuntimeOverlay, n.Id(), err)
			}
		}
		wg.Done()
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}
	for _, n := range nodes {
		nodeChan <- n
	}
	close(nodeChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}
	return nil
}

func BuildSpecificOverlays(nodes []node.Node, allNodes []node.Node, overlayNames []string, workerCount int) error {
	nodeChan := make(chan node.Node, len(nodes))
	errChan := make(chan error, len(nodes))

	var wg sync.WaitGroup
	worker := func() {
		for n := range nodeChan {
			wwlog.Info("Building overlay for %s: %v", n.Id(), overlayNames)
			for _, overlayName := range overlayNames {
				err := BuildOverlay(n, allNodes, "", []string{overlayName})
				if err != nil {
					errChan <- fmt.Errorf("could not build overlay %s for node %s: %w", overlayName, n.Id(), err)
				}
			}
		}
		wg.Done()
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}
	for _, n := range nodes {
		nodeChan <- n
	}
	close(nodeChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}
	return nil
}

/*
Build overlay for the host, so no argument needs to be given
*/
func BuildHostOverlay() error {
	hostname, _ := os.Hostname()
	hostData := node.NewNode(hostname)
	wwlog.Info("Building overlay for %s: host", hostname)
	hostdir, err := Get("host")
	if err != nil {
		return err
	}
	stats, err := os.Stat(hostdir.Rootfs())
	if err != nil {
		return fmt.Errorf("could not build host overlay: %w ", err)
	}
	if !(stats.Mode() == os.FileMode(0o750|os.ModeDir) || stats.Mode() == os.FileMode(0o700|os.ModeDir)) {
		wwlog.SecWarn("Permissions of host overlay dir %s are %s (750 is considered as secure)", hostdir.Rootfs(), stats.Mode())
	}
	registry, err := node.New()
	if err != nil {
		return err
	}
	var allNodes []node.Node
	allNodes, err = registry.FindAllNodes()
	if err != nil {
		return err
	}
	return BuildOverlayIndir(hostData, allNodes, []string{"host"}, "/")
}

/*
Get all overlays present in warewulf
*/
func FindOverlays() (overlayList []string) {
	dotfilecheck, _ := regexp.Compile(`^\..*`)
	controller := config.Get()
	var files []fs.DirEntry
	if distfiles, err := os.ReadDir(controller.Paths.DistributionOverlaydir()); err != nil {
		wwlog.Warn("error reading overlays from %s: %s", controller.Paths.DistributionOverlaydir(), err)
	} else {
		files = append(files, distfiles...)
	}
	if sitefiles, err := os.ReadDir(controller.Paths.SiteOverlaydir()); err != nil {
		wwlog.Warn("error reading overlays from %s: %s", controller.Paths.SiteOverlaydir(), err)
	} else {
		files = append(files, sitefiles...)
	}
	for _, file := range files {
		wwlog.Debug("Evaluating overlay source: %s", file.Name())
		isdotfile := dotfilecheck.MatchString(file.Name())

		if file.IsDir() && !isdotfile && !util.InSlice(overlayList, file.Name()) {
			overlayList = append(overlayList, file.Name())
		}
	}
	return overlayList
}

/*
Build the given overlays for a node and create an image for them
*/
func BuildOverlay(nodeConf node.Node, allNodes []node.Node, context string, overlayNames []string) error {
	if len(overlayNames) == 0 && context == "" {
		return nil
	}

	// create the dir where the overlay images will reside
	var name string
	if context != "" {
		name = fmt.Sprintf("%s %s overlay", nodeConf.Id(), context)
	} else {
		name = fmt.Sprintf("%s overlay/%v", nodeConf.Id(), overlayNames)
	}
	overlayImage := Image(nodeConf.Id(), context, overlayNames)
	overlayImageDir := path.Dir(overlayImage)

	err := os.MkdirAll(overlayImageDir, 0o750)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %s: %w", name, overlayImageDir, err)
	}

	wwlog.Debug("Created directory for %s: %s", name, overlayImageDir)

	buildDir, err := os.MkdirTemp(os.TempDir(), ".wwctl-overlay-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory for %s: %w", name, err)
	}
	defer os.RemoveAll(buildDir)

	wwlog.Debug("Created temporary directory for %s: %s", name, buildDir)

	err = BuildOverlayIndir(nodeConf, allNodes, overlayNames, buildDir)
	if err != nil {
		return fmt.Errorf("failed to generate files for %s: %w", name, err)
	}

	wwlog.Debug("Generated files for %s", name)

	err = util.BuildFsImage(
		name,
		buildDir,
		overlayImage,
		[]string{"*"},
		[]string{},
		// ignore cross-device files
		true,
		"newc")

	return err
}

var (
	regFile *regexp.Regexp
	regLink *regexp.Regexp
)

func init() {
	regFile = regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
	regLink = regexp.MustCompile(`.*{{\s*/\*\s*softlink\s*["'](.*)["']\s*\*/\s*}}.*`)
}

// Build the given overlays for a node in the given directory.
func BuildOverlayIndir(nodeData node.Node, allNodes []node.Node, overlayNames []string, outputDir string) error {
	if len(overlayNames) == 0 {
		return nil
	}
	if !util.IsDir(outputDir) {
		return fmt.Errorf("output must a be a directory: %s", outputDir)
	}

	if !util.ValidString(strings.Join(overlayNames, ""), "^[a-zA-Z0-9-._:]+$") {
		return fmt.Errorf("overlay names contains illegal characters: %v", overlayNames)
	}

	wwlog.Verbose("Processing node/overlays: %s/%s", nodeData.Id(), strings.Join(overlayNames, ","))
	for _, overlayName := range overlayNames {
		wwlog.Verbose("Building overlay %s for node %s in %s", overlayName, nodeData.Id(), outputDir)
		overlayRootfs, err := Get(overlayName)
		if err != nil {
			return err
		}

		wwlog.Debug("Walking the overlay structure: %s", overlayRootfs.Rootfs())
		err = filepath.Walk(overlayRootfs.Rootfs(), func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error for %s: %w", walkPath, err)
			}
			wwlog.Debug("Found overlay file: %s", walkPath)

			relPath, relErr := filepath.Rel(overlayRootfs.Rootfs(), walkPath)
			if relErr != nil {
				wwlog.Warn("Error computing relative path for %s: %v", walkPath, relErr)
				return relErr
			}
			outputPath := path.Join(outputDir, relPath)

			if info.IsDir() {
				wwlog.Debug("Found directory: %s", walkPath)

				if err = os.MkdirAll(outputPath, info.Mode()); err != nil {
					return fmt.Errorf("could not create directory within overlay: %w", err)
				}
				if err = util.CopyUIDGID(walkPath, outputPath); err != nil {
					return fmt.Errorf("failed setting permissions on overlay directory: %w", err)
				}

				wwlog.Debug("Created directory in overlay: %s", outputPath)

			} else if filepath.Ext(walkPath) == ".ww" {
				originalOutputPath := outputPath
				outputPath := strings.TrimSuffix(outputPath, ".ww")
				tstruct, err := InitStruct(overlayName, nodeData, allNodes)
				if err != nil {
					return fmt.Errorf("failed to initial data for %s: %w", nodeData.Id(), err)
				}
				tstruct.BuildSource = walkPath
				wwlog.Verbose("Evaluating overlay template file: %s", walkPath)

				buffer, backupFile, writeFile, err := RenderTemplateFile(walkPath, tstruct)
				if err != nil {
					return fmt.Errorf("failed to render template %s: %w", walkPath, err)
				}
				if !writeFile {
					return nil
				}
				var fileBuffer bytes.Buffer
				// search for magic file name comment
				fileScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
				fileScanner.Split(ScanLines)
				writingToNamedFile := false
				isLink := false
				for fileScanner.Scan() {
					line := fileScanner.Text()
					filenameFromTemplate := regFile.FindAllStringSubmatch(line, -1)
					targetFromTemplate := regLink.FindAllStringSubmatch(line, -1)
					if len(targetFromTemplate) != 0 {
						target := targetFromTemplate[0][1]
						wwlog.Debug("Creating soft link %s -> %s", outputPath, target)
						err := os.Symlink(target, outputPath)
						if err != nil {
							return fmt.Errorf("could not create symlink from template: %w", err)
						} else {
							isLink = true
						}
					} else if len(filenameFromTemplate) != 0 {
						wwlog.Debug("Writing file %s", filenameFromTemplate[0][1])
						if writingToNamedFile && !isLink {
							err = CarefulWriteBuffer(outputPath, fileBuffer, backupFile, info.Mode())
							if err != nil {
								return fmt.Errorf("could not write file from template: %w", err)
							}
							err = util.CopyUIDGID(walkPath, outputPath)
							if err != nil {
								return fmt.Errorf("failed setting permissions on template output file: %w", err)
							}
							fileBuffer.Reset()
						}
						if path.IsAbs(filenameFromTemplate[0][1]) {
							outputPath = filenameFromTemplate[0][1]
							// Create parent directory for absolute paths
							parentDir := path.Dir(outputPath)
							sourceDirInfo, err := os.Stat(path.Dir(walkPath))
							if err != nil {
								return fmt.Errorf("could not stat source directory: %w", err)
							}
							if err := os.MkdirAll(parentDir, sourceDirInfo.Mode()); err != nil {
								return fmt.Errorf("could not create parent directory for absolute path: %w", err)
							}
						} else {
							outputPath = path.Join(path.Dir(originalOutputPath), filenameFromTemplate[0][1])
						}
						writingToNamedFile = true
						isLink = false
					} else {
						if _, err = fileBuffer.WriteString(line); err != nil {
							return fmt.Errorf("could not write to template buffer: %w", err)
						}
					}
				}
				if !isLink {
					err = CarefulWriteBuffer(outputPath, fileBuffer, backupFile, info.Mode())
					if err != nil {
						return fmt.Errorf("could not write file from template: %w", err)
					}
					err = util.CopyUIDGID(walkPath, outputPath)
					if err != nil {
						return fmt.Errorf("failed setting permissions on template output file: %w", err)
					}
					wwlog.Debug("Wrote template file into overlay: %s", outputPath)
				}

			} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				wwlog.Debug("Found symlink %s", walkPath)
				target, err := os.Readlink(walkPath)
				if err != nil {
					return fmt.Errorf("failed reading symlink: %w", err)
				}
				if util.IsFile(outputPath) {
					backupPath := outputPath + ".wwbackup"
					if !util.IsFile(backupPath) {
						wwlog.Debug("Output file already exists: moving to backup file")
						if err = os.Rename(outputPath, backupPath); err != nil {
							return fmt.Errorf("failed renaming to backup file: %w", err)
						}
					} else {
						wwlog.Debug("%s exists, keeping the backup file", backupPath)
						if err = os.Remove(outputPath); err != nil {
							return fmt.Errorf("failed removing existing file: %w", err)
						}
					}
				}
				if err = os.Symlink(target, outputPath); err != nil {
					return fmt.Errorf("failed creating symlink: %w", err)
				}
				wwlog.Debug("Created symlink file: %s", outputPath)
			} else {
				if err := util.CopyFile(walkPath, outputPath); err != nil {
					return fmt.Errorf("could not copy file into overlay: %w", err)
				}
				wwlog.Debug("Copied overlay file: %s", outputPath)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to build overlay image directory: %w", err)
		}
	}

	return nil
}

/*
Writes buffer to the destination file. If wwbackup is set a wwbackup will be created.
*/
func CarefulWriteBuffer(destFile string, buffer bytes.Buffer, backupFile bool, perm fs.FileMode) error {
	wwlog.Debug("Trying to careful write file (%d bytes): %s", buffer.Len(), destFile)
	if backupFile {
		if !util.IsFile(destFile+".wwbackup") && util.IsFile(destFile) {
			err := util.CopyFile(destFile, destFile+".wwbackup")
			if err != nil {
				return fmt.Errorf("failed to create backup: %s -> %s.wwbackup %w", destFile, destFile, err)
			}
		}
	}
	w, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("could not open new file for template %w", err)
	}
	defer w.Close()
	_, err = buffer.WriteTo(w)
	return err
}

/*
Parses the template with the given filename, variables must be in data. Returns the
parsed template as bytes.Buffer, and the bool variables for backupFile and writeFile.
If something goes wrong an error is returned.
*/
func RenderTemplateFile(fileName string, data TemplateStruct) (
	buffer bytes.Buffer,
	backupFile bool,
	writeFile bool,
	err error,
) {
	backupFile = true
	writeFile = true
	// Build our FuncMap
	funcMap := template.FuncMap{
		"Include":      templateFileInclude,
		"IncludeFrom":  templateImageFileInclude,
		"IncludeBlock": templateFileBlock,
		"ImportLink":   importSoftlink,
		"basename":     path.Base,
		"inc":          func(i int) int { return i + 1 },
		"dec":          func(i int) int { return i - 1 },
		"file":         func(str string) string { return fmt.Sprintf("{{ /* file \"%s\" */ }}", str) },
		"softlink":     softlink,
		"readlink":     filepath.EvalSymlinks,
		"IgnitionJson": func() string {
			return createIgnitionJson(data.ThisNode)
		},
		"abort": func() string {
			wwlog.Debug("abort file called in %s", fileName)
			writeFile = false
			return ""
		},
		"nobackup": func() string {
			wwlog.Debug("not backup for %s", fileName)
			backupFile = false
			return ""
		},
		"UniqueField":       UniqueField,
		"SystemdEscape":     unit.UnitNameEscape,
		"SystemdEscapePath": unit.UnitNamePathEscape,
	}

	// Merge sprig.FuncMap with our FuncMap
	for key, value := range sprig.TxtFuncMap() {
		funcMap[key] = value
	}

	// Create the template with the merged FuncMap
	tmpl, err := template.New(path.Base(fileName)).Option("missingkey=default").Funcs(funcMap).ParseGlob(fileName)
	if err != nil {
		err = fmt.Errorf("could not parse template %s: %w", fileName, err)
		return
	}

	err = tmpl.Execute(&buffer, data)
	if err != nil {
		err = fmt.Errorf("could not execute template: %w", err)
		return
	}
	return
}

// Simple version of ScanLines, but include the line break
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// Get all the files as a string slice for a given overlay
func (overlay Overlay) GetFiles() (files []string, err error) {
	err = filepath.Walk(overlay.Rootfs(), func(path string, info fs.FileInfo, err error) error {
		if util.IsFile(path) {
			files = append(files, strings.TrimPrefix(path, overlay.Rootfs()))
		}
		return nil
	})
	return
}
