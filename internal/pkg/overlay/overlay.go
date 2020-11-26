package overlay

import (
	"bytes"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"
)


func OverlayBuild(nodeList []node.NodeInfo, overlayType string) error {
	config := config.New()

	for _, n := range nodeList {
		wwlog.Printf(wwlog.DEBUG, "Processing overlay for node: %s\n", n.Fqdn)

		if util.IsDir(config.VnfsImage(vnfs.CleanName(n.Vnfs))) == false {
			wwlog.Printf(wwlog.WARN, "'%s', VNFS not built: %s\n", n.Fqdn, n.Vnfs)
//			continue
		}

		var files []node.OverlayEntry

		tmpDir, err := ioutil.TempDir(os.TempDir(), ".wwctl-runtime-overlay-")
		if err != nil {
			continue
		}

		if overlayType == "system" {
			wwlog.Printf(wwlog.VERBOSE, "Generating system overlay for node: %s\n", n.Fqdn)
			files = n.SystemOverlay
		} else if overlayType == "runtime" {
			wwlog.Printf(wwlog.VERBOSE, "Generating runtime overlay for node: %s\n", n.Fqdn)
			files = n.RuntimeOverlay
		} else {
			wwlog.Printf(wwlog.ERROR, "overlayType requested is not supported: %s\n", overlayType)
			os.Exit(1)
		}

		for _, file := range files {
			fullPath := path.Join(tmpDir, file.Path)

			mode := file.Mode
			if mode == 0 {
				mode = 0755
			}

			if file.Dir == true {
				wwlog.Printf(wwlog.DEBUG, "Creating overlay directory: %s\n", file.Path)
				err := os.MkdirAll(fullPath, os.FileMode(mode))
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Could not create parent directory in overlay: %s\n", file.Path)
					continue
				}
			} else if file.File == true {
				wwlog.Printf(wwlog.DEBUG, "Creating overlay file: %s\n", file.Path)

				err := os.MkdirAll(path.Dir(fullPath), 0755)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Could not create parent directory for overlay file: %s\n", file.Dir)
					continue
				}

				write, err := os.OpenFile(path.Join(tmpDir, file.Path), os.O_RDWR|os.O_CREATE, os.FileMode(mode))
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					continue
				}

				if file.Source != "" {
					tmpl, _ := template.New("string").Parse(file.Source)
					var tpl bytes.Buffer
					_ = tmpl.Execute(&tpl, n)
					string := tpl.String()

					source, err := os.Open(string)
					if err != nil {
						wwlog.Printf(wwlog.ERROR, "%s\n", err)
						source.Close()
						continue
					}
					_, err = io.Copy(write, source)
					if err != nil {
						wwlog.Printf(wwlog.ERROR, "%s\n", err)
						source.Close()
						continue
					}
					source.Close()
				}

				for _, s := range file.Sources {
					tmpl, _ := template.New("string").Parse(s)
					var tpl bytes.Buffer
					_ = tmpl.Execute(&tpl, n)
					string := tpl.String()

					source, err := os.Open(string)
					if err != nil {
						wwlog.Printf(wwlog.WARN, "%s\n", err)
						source.Close()
						continue
					}
					_, err = io.Copy(write, source)
					if err != nil {
						wwlog.Printf(wwlog.WARN, "%s\n", err)
						source.Close()
						continue
					}
					source.Close()
				}

				write.Close()
			} else if file.Link == true {
				wwlog.Printf(wwlog.DEBUG, "Creating overlay link: %s\n", file.Path)

				if file.Source != "" {
					err := os.MkdirAll(path.Dir(fullPath), 0755)
					if err != nil {
						wwlog.Printf(wwlog.ERROR, "Could not create parent directory for overlay file: %s\n", file.Path)
						continue
					}
					err = os.Symlink(file.Source, fullPath)
					if err != nil {
						wwlog.Printf(wwlog.WARN, "Could not create overlay symlink: %s\n", err)
						continue
					}
				} else {
					wwlog.Printf(wwlog.WARN, "Could not create link, no source defined: %s\n", file.Path)
					continue
				}
			} else if file.Template == true {
				wwlog.Printf(wwlog.DEBUG, "Creating overlay file from template: %s\n", file.Path)

				err := os.MkdirAll(path.Dir(fullPath), 0755)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "Could not create parent directory for overlay file: %s\n", file.Path)
					continue
				}

				if file.Source != "" {
					tmpl, err := template.New(path.Base(file.Source)).Funcs(template.FuncMap{
						"Include":         templateFileInclude,
						"IncludeFromVnfs": templateVnfsFileInclude,
					}).ParseFiles(file.Source)
					if err != nil {
						wwlog.Printf(wwlog.ERROR, "%s\n", err)
						continue
					}

					w, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, os.FileMode(mode))
					if err != nil {
						w.Close()
						wwlog.Printf(wwlog.ERROR, "%s\n", err)
						continue
					}

					err = tmpl.Execute(w, n)
					if err != nil {
						w.Close()
						wwlog.Printf(wwlog.ERROR, "%s\n", err)
						continue
					}
					w.Close()
				}

			} else {
				wwlog.Printf(wwlog.ERROR, "Overlay file type error: %s\n", file.Path)
			}

		}

		var overlayFile string
		if overlayType == "system" {
			overlayFile = config.SystemOverlayImage(n.Fqdn)
		} else if overlayType == "runtime" {
			overlayFile = config.RuntimeOverlayImage(n.Fqdn)
		}

		err = os.MkdirAll(path.Dir(overlayFile), 0755)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not create parent directory for overlay file: %s\n", overlayFile)
			continue
		}


		wwlog.Printf(wwlog.VERBOSE, "Finished generating overlay directory for: %s\n", n.Fqdn)

		os.Chmod(tmpDir, 0755)
		cmd := fmt.Sprintf("cd \"%s\"; find . | cpio --quiet -o -H newc -F \"%s\"", tmpDir, overlayFile)
		wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not generate runtime image overlay: %s\n", err)
			os.RemoveAll(tmpDir)
			continue
		}
		wwlog.Printf(wwlog.INFO, "%-35s: Done\n", n.Fqdn)

		wwlog.Printf(wwlog.DEBUG, "Removing temporary directory: %s\n", tmpDir)
		os.RemoveAll(tmpDir)

	}

	return nil
}
