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
	"syscall"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var (
	ErrDoesNotExist = errors.New("overlay does not exist")
)

/*
Build all overlays for a node
*/
func BuildAllOverlays(nodes []node.Node) error {
	for _, n := range nodes {

		sysOverlays := n.SystemOverlay
		wwlog.Info("Building system overlays for %s: [%s]", n.Id(), strings.Join(sysOverlays, ", "))
		err := BuildOverlay(n, "system", sysOverlays)
		if err != nil {
			return fmt.Errorf("could not build system overlays %v for node %s: %w", sysOverlays, n.Id(), err)
		}
		runOverlays := n.RuntimeOverlay
		wwlog.Info("Building runtime overlays for %s: [%s]", n.Id(), strings.Join(runOverlays, ", "))
		err = BuildOverlay(n, "runtime", runOverlays)
		if err != nil {
			return fmt.Errorf("could not build runtime overlays %v for node %s: %w", runOverlays, n.Id(), err)
		}

	}
	return nil
}

// TODO: Add an Overlay Delete for both sourcedir and image

func BuildSpecificOverlays(nodes []node.Node, overlayNames []string) error {
	for _, n := range nodes {
		wwlog.Info("Building overlay for %s: %v", n, overlayNames)
		for _, overlayName := range overlayNames {
			err := BuildOverlay(n, "", []string{overlayName})
			if err != nil {
				return fmt.Errorf("could not build overlay %s for node %s: %w", overlayName, n.Id(), err)
			}
		}
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
	hostdir, _ := OverlaySourceDir("host")
	stats, err := os.Stat(hostdir)
	if err != nil {
		return fmt.Errorf("could not build host overlay: %w ", err)
	}
	if !(stats.Mode() == os.FileMode(0750|os.ModeDir) || stats.Mode() == os.FileMode(0700|os.ModeDir)) {
		wwlog.SecWarn("Permissions of host overlay dir %s are %s (750 is considered as secure)", hostdir, stats.Mode())
	}
	return BuildOverlayIndir(hostData, []string{"host"}, "/")
}

/*
Get all overlays present in warewulf
*/
func FindOverlays() (overlayList []string, err error) {
	dotfilecheck, _ := regexp.Compile(`^\..*`)
	controller := warewulfconf.Get()
	files, err := os.ReadDir(controller.Paths.WWOverlaydir)
	if err != nil {
		return overlayList, fmt.Errorf("could not get list of distribution overlays: %w", err)
	}
	sitefiles, err := os.ReadDir(path.Join(controller.Paths.Sysconfdir, "overlays"))
	if err == nil { // we don't care if there are no site overlays
		files = append(files, sitefiles...)
	}
	for _, file := range files {
		wwlog.Debug("Evaluating overlay source: %s", file.Name())
		isdotfile := dotfilecheck.MatchString(file.Name())

		if (file.IsDir()) && !(isdotfile) {
			overlayList = append(overlayList, file.Name())
		}
	}

	return overlayList, nil
}

/*
Creates an empty overlay
*/
func OverlayInit(overlayName string) error {
	controller := warewulfconf.Get()
	overlayPath := path.Join(controller.Paths.Sysconfdir, "overlays", overlayName)
	if util.IsDir(overlayPath) {
		return fmt.Errorf("overlay already exists: %s", overlayName)
	}

	err := os.MkdirAll(path.Join(overlayPath, "rootfs"), 0755)

	return err
}

/*
Build the given overlays for a node and create a Image for them
*/
func BuildOverlay(nodeConf node.Node, context string, overlayNames []string) error {
	if len(overlayNames) == 0 && context == "" {
		return nil
	}

	// create the dir where the overlay images will reside
	name := fmt.Sprintf("overlay %s/%v", nodeConf.Id(), overlayNames)
	overlayImage := OverlayImage(nodeConf.Id(), context, overlayNames)
	overlayImageDir := path.Dir(overlayImage)

	err := os.MkdirAll(overlayImageDir, 0750)
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

	err = BuildOverlayIndir(nodeConf, overlayNames, buildDir)
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

/*
Build the given overlays for a node in the given directory. If the given does not
exists it will be created.
*/
func BuildOverlayIndir(nodeData node.Node, overlayNames []string, outputDir string) error {
	if len(overlayNames) == 0 {
		return nil
	}
	if !util.IsDir(outputDir) {
		return errors.Errorf("output must a be a directory: %s", outputDir)
	}

	if !util.ValidString(strings.Join(overlayNames, ""), "^[a-zA-Z0-9-._:]+$") {
		return errors.Errorf("overlay names contains illegal characters: %v", overlayNames)
	}

	// Temporarily set umask to 0000, so directories in the overlay retain permissions
	defer syscall.Umask(syscall.Umask(0))

	wwlog.Verbose("Processing node/overlay: %s/%s", nodeData.Id(), strings.Join(overlayNames, "-"))
	for _, overlayName := range overlayNames {
		wwlog.Verbose("Building overlay %s for node %s in %s", overlayName, nodeData.Id(), outputDir)
		overlaySourceDir, _ := OverlaySourceDir(overlayName)
		wwlog.Debug("Changing directory to OverlayDir: %s", overlaySourceDir)
		err := os.Chdir(overlaySourceDir)
		if err != nil {
			return fmt.Errorf("directory: %s name: %s err: %w", overlaySourceDir, overlayName, ErrDoesNotExist)
		}

		wwlog.Verbose("Walking the overlay structure: %s", overlaySourceDir)
		err = filepath.Walk(".", func(location string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error for %s: %w", location, err)
			}

			wwlog.Debug("Found overlay file: %s", location)

			if info.IsDir() {
				wwlog.Debug("Found directory: %s", location)

				err = os.MkdirAll(path.Join(outputDir, location), info.Mode())
				if err != nil {
					return fmt.Errorf("could not create directory within overlay: %w", err)
				}
				err = util.CopyUIDGID(location, path.Join(outputDir, location))
				if err != nil {
					return fmt.Errorf("failed setting permissions on overlay directory: %w", err)
				}

				wwlog.Debug("Created directory in overlay: %s", location)

			} else if filepath.Ext(location) == ".ww" {
				tstruct, err := InitStruct(nodeData)
				if err != nil {
					return fmt.Errorf("failed to initial data for %s: %w", nodeData.Id(), err)
				}
				tstruct.BuildSource = path.Join(overlaySourceDir, location)
				wwlog.Verbose("Evaluating overlay template file: %s", location)
				destFile := strings.TrimSuffix(location, ".ww")

				buffer, backupFile, writeFile, err := RenderTemplateFile(location, tstruct)
				if err != nil {
					return fmt.Errorf("failed to render template %s: %w", location, err)
				}
				if writeFile {
					destFileName := destFile
					var fileBuffer bytes.Buffer
					// search for magic file name comment
					fileScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
					fileScanner.Split(ScanLines)
					regFile := regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
					regLink := regexp.MustCompile(`.*{{\s*/\*\s*softlink\s*["'](.*)["']\s*\*/\s*}}.*`)
					foundFileComment := false
					for fileScanner.Scan() {
						line := fileScanner.Text()
						filenameFromTemplate := regFile.FindAllStringSubmatch(line, -1)
						softlinkFromTemplate := regLink.FindAllStringSubmatch(line, -1)
						if len(softlinkFromTemplate) != 0 {
							wwlog.Debug("Creating soft link %s -> %s", destFileName, softlinkFromTemplate[0][1])
							return os.Symlink(softlinkFromTemplate[0][1], path.Join(outputDir, destFileName))
						} else if len(filenameFromTemplate) != 0 {
							wwlog.Debug("Writing file %s", filenameFromTemplate[0][1])
							if foundFileComment {
								err = CarefulWriteBuffer(path.Join(outputDir, destFileName),
									fileBuffer, backupFile, info.Mode())
								if err != nil {
									return fmt.Errorf("could not write file from template: %w", err)
								}
								err = util.CopyUIDGID(location, path.Join(outputDir, destFileName))
								if err != nil {
									return fmt.Errorf("failed setting permissions on template output file: %w", err)
								}
								fileBuffer.Reset()
							}
							destFileName = path.Join(path.Dir(destFile), filenameFromTemplate[0][1])
							foundFileComment = true
						} else {
							_, _ = fileBuffer.WriteString(line)
						}
					}
					err = CarefulWriteBuffer(path.Join(outputDir, destFileName), fileBuffer, backupFile, info.Mode())
					if err != nil {
						return fmt.Errorf("could not write file from template: %w", err)
					}
					err = util.CopyUIDGID(location, path.Join(outputDir, destFileName))
					if err != nil {
						return fmt.Errorf("failed setting permissions on template output file: %w", err)
					}

					wwlog.Debug("Wrote template file into overlay: %s", destFile)

					//		} else if b, _ := regexp.MatchString(`\.ww[a-zA-Z0-9\-\._]*$`, location); b {
					//			wwlog.Debug("Ignoring WW template file: %s", location)
				}
			} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				wwlog.Debug("Found symlink %s", location)
				destination, err := os.Readlink(location)
				if err != nil {
					wwlog.ErrorExc(err, "")
				}
				if util.IsFile(path.Join(outputDir, location)) {
					if !util.IsFile(path.Join(outputDir, location+".wwbackup")) {
						wwlog.Debug("Target exists, creating backup file")
						err = os.Rename(path.Join(outputDir, location), path.Join(outputDir, location+".wwbackup"))
					} else {
						wwlog.Debug("%s exists, keeping the backup file", path.Join(outputDir, location+".wwbackup"))
						err = os.Remove(path.Join(outputDir, location))
					}
					if err != nil {
						wwlog.ErrorExc(err, "")
					}
				}
				err = os.Symlink(destination, path.Join(outputDir, location))
				if err != nil {
					wwlog.ErrorExc(err, "")
				}
			} else {
				err := util.CopyFile(location, path.Join(outputDir, location))
				if err == nil {
					wwlog.Debug("Copied file into overlay: %s", location)
				} else {
					return fmt.Errorf("could not copy file into overlay: %w", err)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to build overlay working directory: %w", err)
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
	err error) {
	backupFile = true
	writeFile = true
	// Build our FuncMap
	funcMap := template.FuncMap{
		"Include":      templateFileInclude,
		"IncludeFrom":  templateContainerFileInclude,
		"IncludeBlock": templateFileBlock,
		"ImportLink":   importSoftlink,
		"basename":     path.Base,
		"inc":          func(i int) int { return i + 1 },
		"dec":          func(i int) int { return i - 1 },
		"file":         func(str string) string { return fmt.Sprintf("{{ /* file \"%s\" */ }}", str) },
		"softlink":     softlink,
		"readlink":     filepath.EvalSymlinks,
		"IgnitionJson": func() string {
			str := createIgnitionJson(data.ThisNode)
			if str != "" {
				return str
			}
			writeFile = false
			return ""
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
func OverlayGetFiles(name string) (files []string, err error) {
	baseDir, _ := OverlaySourceDir(name)
	if !util.IsDir(baseDir) {
		err = fmt.Errorf("overlay %s doesn't exist", name)
		return
	}
	err = filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if util.IsFile(path) {
			files = append(files, strings.TrimPrefix(path, baseDir))
		}
		return nil
	})
	return
}
