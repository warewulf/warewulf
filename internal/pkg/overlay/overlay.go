package overlay

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
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
		wwlog.Printf(wwlog.INFO, "Building system overlays for %s: [%s]\n", n.Id.Get(), strings.Join(sysOverlays, ", "))
		err := BuildOverlay(n, sysOverlays)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not build system overlays %v for nide %s\n", sysOverlays, n.Id.Get()))
		}
		runOverlays := n.RuntimeOverlay.GetSlice()
		wwlog.Printf(wwlog.INFO, "Building runtime overlays for %s: [%s]\n", n.Id.Get(), strings.Join(runOverlays, ", "))
		err = BuildOverlay(n, runOverlays)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not build runtime overlays %v for nide %s\n", runOverlays, n.Id.Get()))
		}

	}
	return nil
}

// TODO: Add an Overlay Delete for both sourcedir and image

func BuildSpecificOverlays(nodes []node.NodeInfo, overlayName string) error {
	for _, n := range nodes {

		wwlog.Printf(wwlog.INFO, "Building overlay for %s: %s\n", n.Id.Get(), overlayName)
		err := BuildOverlay(n, []string{overlayName})
		if err != nil {
			return errors.Wrap(err, "could not build overlay "+n.Id.Get()+"/"+overlayName+".img")
		}

	}
	return nil
}

/*
Build overlay for the host, so no argument needs to be given
*/
func BuildHostOverlay() error {
	var host node.NodeInfo
	var idEntry node.Entry
	hostname, _ := os.Hostname()
	wwlog.Printf(wwlog.INFO, "Building overlay for %s: host\n", hostname)
	idEntry.Set(hostname)
	host.Id = idEntry
	return BuildOverlayIndir(host, []string{"host"}, "/")
}

/*
Get all overlays present in warewulf
*/
func FindOverlays() ([]string, error) {
	var ret []string
	var files []os.FileInfo

	files, err := ioutil.ReadDir(OverlaySourceTopDir())
	if err != nil {
		return ret, errors.Wrap(err, "could not get list of overlays")
	}

	for _, file := range files {
		wwlog.Printf(wwlog.DEBUG, "Evaluating overlay source: %s\n", file.Name())
		if file.IsDir() {
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
func BuildOverlay(nodeInfo node.NodeInfo, overlayNames []string) error {
	// create the dir where the overlay images will reside
	overlayImage := OverlayImage(nodeInfo.Id.Get(), overlayNames)
	overlayImageDir := path.Dir(overlayImage)
	err := os.MkdirAll(overlayImageDir, 0755)
	if err == nil {
		wwlog.Printf(wwlog.DEBUG, "Created parent directory for Overlay Images: %s\n", overlayImageDir)
	} else {
		return errors.Wrap(err, "could not create overlay image directory")
	}

	outputDir, err := ioutil.TempDir(os.TempDir(), ".wwctl-overlay-")
	if err == nil {
		wwlog.Printf(wwlog.DEBUG, "Creating temporary directory for overlay files: %s\n", outputDir)
	} else {
		return errors.Wrap(err, "could not create overlay temporary directory")
	}
	err = BuildOverlayIndir(nodeInfo, overlayNames, outputDir)
	if err != nil {
		wwlog.Printf(wwlog.WARN, "Got following error when building overlay: %s\n", err)
	}

	wwlog.Printf(wwlog.DEBUG, "Finished generating overlay working directory for: %s/%v\n", nodeInfo.Id.Get(), overlayNames)
	compressor, err := exec.LookPath("pigz")
	if err != nil {
		wwlog.Printf(wwlog.DEBUG, "Could not locate PIGZ, using GZIP\n")
		compressor = "gzip"
	} else {
		wwlog.Printf(wwlog.DEBUG, "Using PIGZ to compress the overlay: %s\n", compressor)
	}

	cmd := fmt.Sprintf("cd \"%s\"; find . | cpio --quiet -o -H newc | %s -c > \"%s\"", outputDir, compressor, overlayImage)

	wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
	err = exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		return errors.Wrap(err, "could not generate compressed runtime image overlay")
	}
	wwlog.Printf(wwlog.VERBOSE, "Completed building overlay image: %s\n", overlayImage)

	wwlog.Printf(wwlog.DEBUG, "Removing temporary directory: %s\n", outputDir)
	os.RemoveAll(outputDir)
	return nil
}

/*
Build the given overlays for a node in the given directory. If the given does not
exists it will be created.
*/
func BuildOverlayIndir(nodeInfo node.NodeInfo, overlayNames []string, outputDir string) error {
	if len(overlayNames) == 0 {
		return errors.New("At least one valid overlay is needed to build for a node\n")
	}
	if !util.IsDir(outputDir) {
		return errors.New(fmt.Sprintf("output %s must a be a directory\n", outputDir))
	}
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	allNodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if !util.ValidString(strings.Join(overlayNames, ""), "^[a-zA-Z0-9-._:]+$") {
		return errors.New(fmt.Sprintf("overlay names contains illegal characters: %v", overlayNames))
	}
	wwlog.Printf(wwlog.VERBOSE, "Processing node/overlay: %s/%s\n", nodeInfo.Id.Get(), strings.Join(overlayNames, "-"))
	var tstruct TemplateStruct
	tstruct.Id = nodeInfo.Id.Get()
	tstruct.Hostname = nodeInfo.Id.Get()
	tstruct.Id = nodeInfo.Id.Get()
	tstruct.Hostname = nodeInfo.Id.Get()
	tstruct.ClusterName = nodeInfo.ClusterName.Get()
	tstruct.Container = nodeInfo.ContainerName.Get()
	tstruct.Kernel.Version = nodeInfo.Kernel.Override.Get()
	tstruct.Kernel.Override = nodeInfo.Kernel.Override.Get()
	tstruct.Kernel.Args = nodeInfo.Kernel.Args.Get()
	tstruct.Init = nodeInfo.Init.Get()
	tstruct.Root = nodeInfo.Root.Get()
	tstruct.Ipmi.Ipaddr = nodeInfo.Ipmi.Ipaddr.Get()
	tstruct.Ipmi.Netmask = nodeInfo.Ipmi.Netmask.Get()
	tstruct.Ipmi.Port = nodeInfo.Ipmi.Port.Get()
	tstruct.Ipmi.Gateway = nodeInfo.Ipmi.Gateway.Get()
	tstruct.Ipmi.UserName = nodeInfo.Ipmi.UserName.Get()
	tstruct.Ipmi.Password = nodeInfo.Ipmi.Password.Get()
	tstruct.Ipmi.Interface = nodeInfo.Ipmi.Interface.Get()
	tstruct.Ipmi.Write = nodeInfo.Ipmi.Write.GetB()
	tstruct.RuntimeOverlay = nodeInfo.RuntimeOverlay.Print()
	tstruct.SystemOverlay = nodeInfo.SystemOverlay.Print()
	tstruct.NetDevs = make(map[string]*node.NetDevs)
	tstruct.Keys = make(map[string]string)
	tstruct.Tags = make(map[string]string)
	for devname, netdev := range nodeInfo.NetDevs {
		var nd node.NetDevs
		tstruct.NetDevs[devname] = &nd
		tstruct.NetDevs[devname].Device = netdev.Device.Get()
		tstruct.NetDevs[devname].Hwaddr = netdev.Hwaddr.Get()
		tstruct.NetDevs[devname].Ipaddr = netdev.Ipaddr.Get()
		tstruct.NetDevs[devname].Netmask = netdev.Netmask.Get()
		tstruct.NetDevs[devname].Gateway = netdev.Gateway.Get()
		tstruct.NetDevs[devname].Type = netdev.Type.Get()
		tstruct.NetDevs[devname].OnBoot = netdev.OnBoot.Get()
		tstruct.NetDevs[devname].Default = netdev.Default.Get()
		mask := net.IPMask(net.ParseIP(netdev.Netmask.Get()).To4())
		ipaddr := net.ParseIP(netdev.Ipaddr.Get()).To4()
		netaddr := net.IPNet{IP: ipaddr, Mask: mask}
		netPrefix, _ := net.IPMask(net.ParseIP(netdev.Netmask.Get()).To4()).Size()
		tstruct.NetDevs[devname].Prefix = strconv.Itoa(netPrefix)
		tstruct.NetDevs[devname].IpCIDR = netaddr.String()
		tstruct.NetDevs[devname].Ipaddr6 = netdev.Ipaddr6.Get()
		tstruct.NetDevs[devname].Tags = make(map[string]string)
		for key, value := range netdev.Tags {
			tstruct.NetDevs[devname].Tags[key] = value.Get()
		}
	}
	// Backwards compatibility for templates using "Keys"
	for keyname, key := range nodeInfo.Tags {
		tstruct.Keys[keyname] = key.Get()
	}
	for keyname, key := range nodeInfo.Tags {
		tstruct.Tags[keyname] = key.Get()
	}
	tstruct.AllNodes = allNodes
	tstruct.Nfs = *controller.Nfs
	tstruct.Dhcp = *controller.Dhcp
	tstruct.Warewulf = *controller.Warewulf
	tstruct.Ipaddr = controller.Ipaddr
	tstruct.Ipaddr6 = controller.Ipaddr6
	tstruct.Netmask = controller.Netmask
	tstruct.Network = controller.Network
	if controller.Ipaddr6 != "" {
		tstruct.Ipv6 = true
	} else {
		tstruct.Ipv6 = false
	}
	hostname, _ := os.Hostname()
	tstruct.BuildHost = hostname
	dt := time.Now()
	tstruct.BuildTime = dt.Format("01-02-2006 15:04:05 MST")
	for _, overlayName := range overlayNames {
		wwlog.Printf(wwlog.VERBOSE, "Building overlay %s for node %s in %s\n", overlayName, nodeInfo.Id.Get(), outputDir)
		overlaySourceDir := OverlaySourceDir(overlayName)
		wwlog.Printf(wwlog.DEBUG, "Starting to build overlay %s\nChanging directory to OverlayDir: %s\n", overlayName, overlaySourceDir)
		err = os.Chdir(overlaySourceDir)
		if err != nil {
			return errors.Wrap(err, "could not change directory to overlay dir")
		}
		wwlog.Printf(wwlog.DEBUG, "Checking to see if overlay directory exists: %s\n", overlaySourceDir)
		if !util.IsDir(overlaySourceDir) {
			return errors.New("overlay does not exist: " + overlayName)
		}

		wwlog.Printf(wwlog.VERBOSE, "Walking the overlay structure: %s\n", overlaySourceDir)
		err = filepath.Walk(".", func(location string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrap(err, "error for "+location)
			}

			wwlog.Printf(wwlog.DEBUG, "Found overlay file: %s\n", location)

			if info.IsDir() {
				wwlog.Printf(wwlog.DEBUG, "Found directory: %s\n", location)

				err = os.MkdirAll(path.Join(outputDir, location), info.Mode())
				if err != nil {
					return errors.Wrap(err, "could not create directory within overlay")
				}
				err = util.CopyUIDGID(location, path.Join(outputDir, location))
				if err != nil {
					return errors.Wrap(err, "failed setting permissions on overlay directory")
				}

				wwlog.Printf(wwlog.DEBUG, "Created directory in overlay: %s\n", location)

			} else if filepath.Ext(location) == ".ww" {
				tstruct.BuildSource = path.Join(overlaySourceDir, location)
				wwlog.Printf(wwlog.VERBOSE, "Evaluating overlay template file: %s\n", location)
				destFile := strings.TrimSuffix(location, ".ww")
				backupFile := true
				writeFile := true
				tmpl, err := template.New(path.Base(location)).Option("missingkey=default").Funcs(template.FuncMap{
					// TODO: Fix for missingkey=zero
					"Include":      templateFileInclude,
					"IncludeFrom":  templateContainerFileInclude,
					"IncludeBlock": templateFileBlock,
					"inc":          func(i int) int { return i + 1 },
					"dec":          func(i int) int { return i - 1 },
					"file":         func(str string) string { return fmt.Sprintf("{{ /* file \"%s\" */ }}", str) },
					"abort": func() string {
						wwlog.Printf(wwlog.DEBUG, "abort file called in %s\n", location)
						writeFile = false
						return ""
					},
					"nobackup": func() string {
						wwlog.Printf(wwlog.DEBUG, "not backup for %s\n", location)
						backupFile = false
						return ""
					},
					// }).ParseGlob(path.Join(OverlayDir, destFile+".ww*"))
				}).ParseGlob(location)
				if err != nil {
					return errors.Wrap(err, "could not parse template "+location)
				}
				var buffer bytes.Buffer
				err = tmpl.Execute(&buffer, tstruct)
				if err != nil {
					return errors.Wrap(err, "could not execute template")

				}
				if writeFile {
					destFileName := destFile
					var fileBuffer bytes.Buffer
					// search for magic file name comment
					fileScanner := bufio.NewScanner(bytes.NewReader(buffer.Bytes()))
					fileScanner.Split(scanLines)
					reg := regexp.MustCompile(`.*{{\s*/\*\s*file\s*["'](.*)["']\s*\*/\s*}}.*`)
					foundFileComment := false
					for fileScanner.Scan() {
						line := fileScanner.Text()
						filenameFromTemplate := reg.FindAllStringSubmatch(line, -1)
						if len(filenameFromTemplate) != 0 {
							wwlog.Printf(wwlog.DEBUG, "Found multifile comment, new filename %s\n", filenameFromTemplate[0][1])
							if foundFileComment {
								err = carefulWriteBuffer(path.Join(outputDir, destFileName),
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
					err = carefulWriteBuffer(path.Join(outputDir, destFileName), fileBuffer, backupFile, info.Mode())
					if err != nil {
						return errors.Wrap(err, "could not write file from template")
					}
					err = util.CopyUIDGID(location, path.Join(outputDir, destFileName))
					if err != nil {
						return errors.Wrap(err, "failed setting permissions on template output file")
					}

					wwlog.Printf(wwlog.DEBUG, "Wrote template file into overlay: %s\n", destFile)

					//		} else if b, _ := regexp.MatchString(`\.ww[a-zA-Z0-9\-\._]*$`, location); b {
					//			wwlog.Printf(wwlog.DEBUG, "Ignoring WW template file: %s\n", location)
				}
			} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				wwlog.Printf(wwlog.DEBUG, "Found symlink %s\n", location)
				destination, err := os.Readlink(location)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
				}
				err = os.Symlink(destination, path.Join(outputDir, location))
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
				}
			} else {

				err := util.CopyFile(location, path.Join(outputDir, location))
				if err == nil {
					wwlog.Printf(wwlog.DEBUG, "Copied file into overlay: %s\n", location)
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
func carefulWriteBuffer(destFile string, buffer bytes.Buffer, backupFile bool, perm fs.FileMode) error {
	wwlog.Printf(wwlog.DEBUG, "Trying to careful write file %s\n", destFile)
	if backupFile {
		// if !util.IsFile(path.Join(outputDir, destFile+".wwbackup")) && util.IsFile(path.Join(outputDir, destFile)) {
		if !util.IsFile(destFile+".wwbackup") && util.IsFile(destFile) {
			err := util.CopyFile(destFile, destFile+".wwbackup")
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
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

// Simple version of ScanLines, but include the line break
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
