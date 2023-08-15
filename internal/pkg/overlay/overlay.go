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

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

/*

func BuildSystemOverlay(nodeList []node.NodeInfo) error {
	return nil
}

func BuildRuntimeOverlay(nodeList []node.NodeInfo) error {
	return nil
}


func FindSystemOverlays() ([]string, error) {
	return findAllOverlays("system")
}

func FindRuntimeOverlays() ([]string, error) {
	return findAllOverlays("runtime")
}
*/

/*
Build all overlays (runtime and generic) for a node
*/
func BuildAllOverlays(nodes []node.NodeInfo) error {
	for _, n := range nodes {

		sysOverlays := n.SystemOverlay.GetSlice()
		wwlog.Info("Building system overlays for %s: [%s]", n.Id.Get(), strings.Join(sysOverlays, ", "))
		err := BuildOverlay(n, sysOverlays, "system")
		if err != nil {
			return errors.Wrapf(err, "could not build system overlays %v for node %s", sysOverlays, n.Id.Get())
		}
		runOverlays := n.RuntimeOverlay.GetSlice()
		wwlog.Info("Building runtime overlays for %s: [%s]", n.Id.Get(), strings.Join(runOverlays, ", "))
		err = BuildOverlay(n, runOverlays, "runtime")
		if err != nil {
			return errors.Wrapf(err, "could not build runtime overlays %v for node %s", runOverlays, n.Id.Get())
		}

	}
	return nil
}

// TODO: Add an Overlay Delete for both sourcedir and image

func BuildSpecificOverlays(nodes []node.NodeInfo, overlayNames []string) error {
	for _, n := range nodes {
		wwlog.Info("Building overlay for %s: %v", n.Id.Get(), overlayNames)
                for _, overlayName := range overlayNames {
                        err := BuildOverlay(n, []string{overlayName})
		        if err != nil {
			      return errors.Wrapf(err, "could not build overlay %s for node %s", overlayName, n.Id.Get())
                        }
		}

	}
	return nil
}

/*
Build overlay for the host, so no argument needs to be given
*/
func BuildHostOverlay() error {
	host := node.NewInfo()
	hostname, _ := os.Hostname()
	host.Id.Set(hostname)

	wwlog.Info("Building overlay for %s: host", hostname)
	hostdir := OverlaySourceDir("host")
	stats, err := os.Stat(hostdir)
	if err != nil {
		return errors.Wrap(err, "could not build host overlay ")
	}
	if !(stats.Mode() == os.FileMode(0750|os.ModeDir) || stats.Mode() == os.FileMode(0700|os.ModeDir)) {
		wwlog.SecWarn("Permissions of host overlay dir %s are %s (750 is considered as secure)", hostdir, stats.Mode())
	}
	return BuildOverlayIndir(host, []string{"host"}, "/")
}

/*
Get all overlays present in warewulf
*/
func FindOverlays() ([]string, error) {
	var ret []string
	dotfilecheck, _ := regexp.Compile(`^\..*`)

	files, err := os.ReadDir(OverlaySourceTopDir())
	if err != nil {
		return ret, errors.Wrap(err, "could not get list of overlays")
	}

	for _, file := range files {
		wwlog.Debug("Evaluating overlay source: %s", file.Name())
		isdotfile := dotfilecheck.MatchString(file.Name())

		if (file.IsDir()) && !(isdotfile) {
			ret = append(ret, file.Name())
		}
	}

	return ret, nil
}

/*
Creates an empty overlay
*/
func OverlayInit(overlayName string) error {
	path := OverlaySourceDir(overlayName)

	if util.IsDir(path) {
		return errors.New("Overlay already exists: " + overlayName)
	}

	err := os.MkdirAll(path, 0755)

	return err
}

/*
Build the given overlays for a node and create a Image for them
*/
func BuildOverlay(nodeInfo node.NodeInfo, overlayNames []string, img_context ...string) error {
	var context string
	/* Check optional context argument. If missing, default to legacy. */
	if len(img_context) == 0 {
		context = "legacy"
	} else {
		context = img_context[0]
	}

	// create the dir where the overlay images will reside
	name := fmt.Sprintf("overlay %s/%v", nodeInfo.Id.Get(), overlayNames)
	overlayImage := OverlayImage(nodeInfo.Id.Get(), overlayNames, context)
	overlayImageDir := path.Dir(overlayImage)

	err := os.MkdirAll(overlayImageDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory for %s: %s", name, overlayImageDir)
	}

	wwlog.Debug("Created directory for %s: %s", name, overlayImageDir)

	buildDir, err := os.MkdirTemp(os.TempDir(), ".wwctl-overlay-")
	if err != nil {
		return errors.Wrapf(err, "Failed to create temporary directory for %s", name)
	}
	defer os.RemoveAll(buildDir)

	wwlog.Debug("Created temporary directory for %s: %s", name, buildDir)

	err = BuildOverlayIndir(nodeInfo, overlayNames, buildDir)
	if err != nil {
		return errors.Wrapf(err, "Failed to generate files for %s", name)
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
func BuildOverlayIndir(nodeInfo node.NodeInfo, overlayNames []string, outputDir string) error {
	if len(overlayNames) == 0 {
		return errors.New("At least one valid overlay is needed to build for a node")
	}
	if !util.IsDir(outputDir) {
		return errors.Errorf("output must a be a directory: %s", outputDir)
	}

	if !util.ValidString(strings.Join(overlayNames, ""), "^[a-zA-Z0-9-._:]+$") {
		return errors.Errorf("overlay names contains illegal characters: %v", overlayNames)
	}

	// Temporarily set umask to 0000, so directories in the overlay retain permissions
	defer syscall.Umask(syscall.Umask(0))

	wwlog.Verbose("Processing node/overlay: %s/%s", nodeInfo.Id.Get(), strings.Join(overlayNames, "-"))
	for _, overlayName := range overlayNames {
		wwlog.Verbose("Building overlay %s for node %s in %s", overlayName, nodeInfo.Id.Get(), outputDir)
		overlaySourceDir := OverlaySourceDir(overlayName)
		wwlog.Debug("Starting to build overlay %s\nChanging directory to OverlayDir: %s", overlayName, overlaySourceDir)
		err := os.Chdir(overlaySourceDir)
		if err != nil {
			return errors.Wrap(err, "could not change directory to overlay dir")
		}
		wwlog.Debug("Checking to see if overlay directory exists: %s", overlaySourceDir)
		if !util.IsDir(overlaySourceDir) {
			return errors.New("overlay does not exist: " + overlayName)
		}

		wwlog.Verbose("Walking the overlay structure: %s", overlaySourceDir)
		err = filepath.Walk(".", func(location string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrap(err, "error for "+location)
			}

			wwlog.Debug("Found overlay file: %s", location)

			if info.IsDir() {
				wwlog.Debug("Found directory: %s", location)

				err = os.MkdirAll(path.Join(outputDir, location), info.Mode())
				if err != nil {
					return errors.Wrap(err, "could not create directory within overlay")
				}
				err = util.CopyUIDGID(location, path.Join(outputDir, location))
				if err != nil {
					return errors.Wrap(err, "failed setting permissions on overlay directory")
				}

				wwlog.Debug("Created directory in overlay: %s", location)

			} else if filepath.Ext(location) == ".ww" {
				tstruct := InitStruct(&nodeInfo)
				tstruct.BuildSource = path.Join(overlaySourceDir, location)
				wwlog.Verbose("Evaluating overlay template file: %s", location)
				destFile := strings.TrimSuffix(location, ".ww")

				buffer, backupFile, writeFile, err := RenderTemplateFile(location, tstruct)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("Failed to render template %s", location))
				}
				if writeFile {
					destFileName := destFile
					var fileBuffer bytes.Buffer
					// search for magic file name comment
					fileScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
					fileScanner.Split(ScanLines)
					reg := regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
					foundFileComment := false
					for fileScanner.Scan() {
						line := fileScanner.Text()
						filenameFromTemplate := reg.FindAllStringSubmatch(line, -1)
						if len(filenameFromTemplate) != 0 {
							wwlog.Debug("Found multiple comment, new filename %s", filenameFromTemplate[0][1])
							if foundFileComment {
								err = CarefulWriteBuffer(path.Join(outputDir, destFileName),
									fileBuffer, backupFile, info.Mode())
								if err != nil {
									return errors.Wrap(err, "could not write file from template")
								}
								err = util.CopyUIDGID(location, path.Join(outputDir, destFileName))
								if err != nil {
									return errors.Wrap(err, "failed setting permissions on template output file")
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
						return errors.Wrap(err, "could not write file from template")
					}
					err = util.CopyUIDGID(location, path.Join(outputDir, destFileName))
					if err != nil {
						return errors.Wrap(err, "failed setting permissions on template output file")
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
					return errors.Wrap(err, "could not copy file into overlay")
				}
			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to build overlay working directory")
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
				return errors.Wrapf(err, "Failed to create backup: %s -> %s.wwbackup", destFile, destFile)
			}
		}

	}
	w, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return errors.Wrap(err, "could not open new file for template")
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
	tmpl, err := template.New(path.Base(fileName)).Option("missingkey=default").Funcs(template.FuncMap{
		// TODO: Fix for missingkey=zero
		"Include":      templateFileInclude,
		"IncludeFrom":  templateContainerFileInclude,
		"IncludeBlock": templateFileBlock,
		"basename":     path.Base,
		"inc":          func(i int) int { return i + 1 },
		"dec":          func(i int) int { return i - 1 },
		"file":         func(str string) string { return fmt.Sprintf("{{ /* file \"%s\" */ }}", str) },
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
		"split": func(s string, d string) []string {
			return strings.Split(s, d)
		},
		"tr": func(source, old, new string) string {
			return strings.Replace(source, old, new, -1)
		},
		"replace": func(source, old, new string) string {
			return strings.Replace(source, old, new, -1)
		},
		// }).ParseGlob(path.Join(OverlayDir, destFile+".ww*"))
	}).ParseGlob(fileName)
	if err != nil {
		err = errors.Wrap(err, "could not parse template "+fileName)
		return
	}
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		err = errors.Wrap(err, "could not execute template")
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
