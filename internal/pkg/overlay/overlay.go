package overlay

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
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
		wwlog.Info("Building system overlays for %s: [%s]", n.Id.Get(), strings.Join(sysOverlays, ", "))
		err := BuildOverlay(n, sysOverlays)
		if err != nil {
			return errors.Wrapf(err, "could not build system overlays %v for nide %s", sysOverlays, n.Id.Get())
		}
		runOverlays := n.RuntimeOverlay.GetSlice()
		wwlog.Info("Building runtime overlays for %s: [%s]", n.Id.Get(), strings.Join(runOverlays, ", "))
		err = BuildOverlay(n, runOverlays)
		if err != nil {
			return errors.Wrapf(err, "could not build runtime overlays %v for nide %s", runOverlays, n.Id.Get())
		}

	}
	return nil
}

// TODO: Add an Overlay Delete for both sourcedir and image

func BuildSpecificOverlays(nodes []node.NodeInfo, overlayName string) error {
	for _, n := range nodes {

		wwlog.Info("Building overlay for %s: %s", n.Id.Get(), overlayName)
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
	host.Kernel = new(node.KernelEntry)
	host.Ipmi = new(node.IpmiEntry)
	var idEntry node.Entry
	hostname, _ := os.Hostname()
	wwlog.Info("Building overlay for %s: host", hostname)
	idEntry.Set(hostname)
	host.Id = idEntry
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
	var files []os.FileInfo
	dotfilecheck, _ := regexp.Compile(`^\..*`)

	files, err := ioutil.ReadDir(OverlaySourceTopDir())
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
func BuildOverlay(nodeInfo node.NodeInfo, overlayNames []string) error {
	// create the dir where the overlay images will reside
	name := fmt.Sprintf("overlay %s/%v", nodeInfo.Id.Get(), overlayNames)
	overlayImage := OverlayImage(nodeInfo.Id.Get(), overlayNames)
	overlayImageDir := path.Dir(overlayImage)

	err := os.MkdirAll(overlayImageDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory for %s: %s", name, overlayImageDir)
	}

	wwlog.Debug("Created directory for %s: %s", name, overlayImageDir)

	buildDir, err := ioutil.TempDir(os.TempDir(), ".wwctl-overlay-")
	if err != nil {
		return errors.Wrapf(err, "Failed to create temporary directory for %s", name )
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
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.ErrorExc(err, "")
		os.Exit(1)
	}
	nodeDB, err := node.New()
	if err != nil {
		wwlog.ErrorExc(err, "")
		os.Exit(1)
	}
	allNodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.ErrorExc(err, "")
		os.Exit(1)
	}

	if !util.ValidString(strings.Join(overlayNames, ""), "^[a-zA-Z0-9-._:]+$") {
		return errors.Errorf("overlay names contains illegal characters: %v", overlayNames)
	}
	wwlog.Verbose("Processing node/overlay: %s/%s", nodeInfo.Id.Get(), strings.Join(overlayNames, "-"))
	var tstruct TemplateStruct
	tstruct.Kernel = new(node.KernelConf)
	tstruct.Ipmi = new(node.IpmiConf)
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
	tstruct.Ipmi.Write = nodeInfo.Ipmi.Write.Get()
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
	netaddrStruct := net.IPNet{IP: net.ParseIP(controller.Network), Mask: net.IPMask(net.ParseIP(controller.Netmask))}
	tstruct.NetworkCIDR = netaddrStruct.String()
	if controller.Ipaddr6 != "" {
		tstruct.Ipv6 = true
	} else {
		tstruct.Ipv6 = false
	}
	hostname, _ := os.Hostname()
	tstruct.BuildHost = hostname
	dt := time.Now()
	tstruct.BuildTime = dt.Format("01-02-2006 15:04:05 MST")
	tstruct.BuildTimeUnix = strconv.FormatInt(dt.Unix(), 10)
	for _, overlayName := range overlayNames {
		wwlog.Verbose("Building overlay %s for node %s in %s", overlayName, nodeInfo.Id.Get(), outputDir)
		overlaySourceDir := OverlaySourceDir(overlayName)
		wwlog.Debug("Starting to build overlay %s\nChanging directory to OverlayDir: %s", overlayName, overlaySourceDir)
		err = os.Chdir(overlaySourceDir)
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
				tstruct.BuildSource = path.Join(overlaySourceDir, location)
				wwlog.Verbose("Evaluating overlay template file: %s", location)
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
						wwlog.Debug("abort file called in %s", location)
						writeFile = false
						return ""
					},
					"nobackup": func() string {
						wwlog.Debug("not backup for %s", location)
						backupFile = false
						return ""
					},
					"split": func(s string, d string) []string {
						return strings.Split(s, d)
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
							wwlog.Debug("Found multifile comment, new filename %s", filenameFromTemplate[0][1])
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
func carefulWriteBuffer(destFile string, buffer bytes.Buffer, backupFile bool, perm fs.FileMode) error {
	wwlog.Debug("Trying to careful write file (%d bytes): %s", buffer.Len(), destFile )
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
