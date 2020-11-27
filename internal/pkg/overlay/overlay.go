package overlay

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type TemplateStruct struct {
	Fqdn 			string
	Hostname		string
	Groupname		string
	Vnfs 			string
	Netdev	 		map[string]*node.NetDevs
}

func BuildSystemOverlay(nodeList []node.NodeInfo) error {
	return buildOverlay(nodeList, "system")
}

func BuildRuntimeOverlay(nodeList []node.NodeInfo) error {
	return buildOverlay(nodeList, "runtime")
}

func FindSystemOverlays() ([]string, error) {
	return findAllOverlays("system")
}

func FindRuntimeOverlays() ([]string, error) {
	return findAllOverlays("system")
}

func SystemOverlayInit(name string) error {
	return overlayInit(name, "system")
}

func RuntimeOverlayInit(name string) error {
	return overlayInit(name, "runtime")
}




func findAllOverlays(overlayType string) ([]string, error) {
	config := config.New()
	var ret []string
	var files []os.FileInfo
	var err error

	if overlayType == "system" {
		wwlog.Printf(wwlog.DEBUG, "Looking for system overlays...")
		files, err = ioutil.ReadDir(config.SystemOverlayDir())
	} else if overlayType == "runtime" {
		wwlog.Printf(wwlog.DEBUG, "Looking for runtime overlays...")
		files, err = ioutil.ReadDir(config.RuntimeOverlayDir())
	} else {
		wwlog.Printf(wwlog.ERROR, "overlayType requested is not supported: %s\n", overlayType)
		os.Exit(1)
	}

	if err != nil {
		return ret, err
	}

	for _, file := range files {
		wwlog.Printf(wwlog.DEBUG, "Evaluating overlay source: %s\n", file.Name())
		if file.IsDir() == true {
			ret = append(ret, file.Name())
		}
	}

	return ret, nil
}


func overlayInit(name string, overlayType string) error {
	var path string
	config := config.New()

	if overlayType == "system" {
		wwlog.Printf(wwlog.DEBUG, "Looking for system overlays...")
		path = config.SystemOverlaySource(name)
	} else if overlayType == "runtime" {
		wwlog.Printf(wwlog.DEBUG, "Looking for runtime overlays...")
		path = config.RuntimeOverlaySource(name)
	} else {
		wwlog.Printf(wwlog.ERROR, "overlayType requested is not supported: %s\n", overlayType)
		os.Exit(1)
	}

	if util.IsDir(path) == true {
		return errors.New("Overlay already exists: "+name)
	}

	err := os.MkdirAll(path, 0755)

	return err
}






func buildOverlay(nodeList []node.NodeInfo, overlayType string) error {
	config := config.New()

	for _, node := range nodeList {
		var t TemplateStruct
		var OverlayDir string
		var OverlayFile string

		if overlayType == "runtime" {
			OverlayDir = config.RuntimeOverlaySource(node.RuntimeOverlay.Get())
			OverlayFile = config.RuntimeOverlayImage(node.Fqdn.Get())
		} else if overlayType == "system" {
			OverlayDir = config.SystemOverlaySource(node.RuntimeOverlay.Get())
			OverlayFile = config.SystemOverlayImage(node.Fqdn.Get())
		} else {
			wwlog.Printf(wwlog.ERROR, "overlayType requested is not supported: %s\n", overlayType)
			os.Exit(1)
		}

		wwlog.Printf(wwlog.DEBUG, "Processing overlay for node: %s\n", node.Fqdn.Get())

		t.Fqdn = node.Fqdn.Get()
		t.Hostname = node.HostName.Get()
		t.Groupname = node.GroupName.Get()
		t.Vnfs = node.Vnfs.Get()
		t.Netdev = node.NetDevs

		if overlayType == "runtime" && node.RuntimeOverlay.Defined() == false {
			wwlog.Printf(wwlog.WARN, "Undefined runtime overlay, skipping node: %s\n", node.Fqdn.Get())
		}
		if overlayType == "system" && node.SystemOverlay.Defined() == false {
			wwlog.Printf(wwlog.WARN, "Undefined system overlay, skipping node: %s\n", node.Fqdn.Get())
		}

		wwlog.Printf(wwlog.DEBUG, "Checking to see if overlay directory exists: %s\n", OverlayDir)
		if util.IsDir(OverlayDir) == false {
			wwlog.Printf(wwlog.WARN, "%-35s: Skipped (runtime overlay template not found)\n", node.Fqdn.Get())
			continue
		}

		wwlog.Printf(wwlog.DEBUG, "Creating parent directory for OverlayFile: %s\n", path.Dir(OverlayFile))
		err := os.MkdirAll(path.Dir(OverlayFile), 0755)
		if err != nil {
			return err
		}

		wwlog.Printf(wwlog.DEBUG, "Changing directory to OverlayDir: %s\n", OverlayDir)
		err = os.Chdir(OverlayDir)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not chdir() to OverlayDir: %s\n", OverlayDir)
			continue
		}

		wwlog.Printf(wwlog.DEBUG, "Creating temporary directory for overlay files\n")
		tmpDir, err := ioutil.TempDir(os.TempDir(), ".wwctl-overlay-")
		if err != nil {
			return err
		}

		wwlog.Printf(wwlog.DEBUG, "Walking the file system: %s\n", OverlayDir)
		err = filepath.Walk(".", func(location string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				wwlog.Printf(wwlog.DEBUG, "Found directory: %s\n", location)

				err := os.MkdirAll(path.Join(tmpDir, location), info.Mode())
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					return err
				}

			} else if filepath.Ext(location) == ".ww" {
				wwlog.Printf(wwlog.DEBUG, "Found template file: %s\n", location)

				destFile := strings.TrimSuffix(location, ".ww")

				tmpl, err := template.New(path.Base(location)).Funcs(template.FuncMap{
					"Include":         templateFileInclude,
					"IncludeFromVnfs": templateVnfsFileInclude,
				}).ParseGlob(path.Join(OverlayDir, destFile+".ww*"))
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					return err
				}

				w, err := os.OpenFile(path.Join(tmpDir, destFile), os.O_RDWR|os.O_CREATE, info.Mode())
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					return err
				}
				defer w.Close()

				err = tmpl.Execute(w, t)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					return err
				}

			} else if b, _ := regexp.MatchString(`\.ww[a-zA-Z0-9\-\._]*$`, location); b == true {
				wwlog.Printf(wwlog.DEBUG, "Ignoring WW template file: %s\n", location)
			} else {
				wwlog.Printf(wwlog.DEBUG, "Found file: %s\n", location)

				err := util.CopyFile(path.Join(OverlayDir, location), path.Join(tmpDir, location))
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					return err
				}

			}

			return nil
		})


		wwlog.Printf(wwlog.VERBOSE, "Finished generating overlay directory for: %s\n", node.Fqdn.Get())

		cmd := fmt.Sprintf("cd \"%s\"; find . | cpio --quiet -o -H newc -F \"%s\"", tmpDir, OverlayFile)
		wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not generate runtime image overlay: %s\n", err)
			continue
		}
		wwlog.Printf(wwlog.INFO, "%-35s: Done\n", node.Fqdn.Get())

		wwlog.Printf(wwlog.DEBUG, "Removing temporary directory: %s\n", tmpDir)
		os.RemoveAll(tmpDir)

	}

	return nil
}
