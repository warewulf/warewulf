package overlay

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

type TemplateStruct struct {
	Id             string
	Hostname       string
	ClusterName    string
	Container      string
	KernelVersion  string
	KernelOverride string
	KernelArgs     string
	Init           string
	Root           string
	IpmiIpaddr     string
	IpmiNetmask    string
	IpmiPort       string
	IpmiGateway    string
	IpmiUserName   string
	IpmiPassword   string
	IpmiInterface  string
	RuntimeOverlay string
	SystemOverlay  string
	NetDevs        map[string]*node.NetDevs
	Tags           map[string]string
	Keys           map[string]string
	AllNodes       []node.NodeInfo
	NFSMounts      []string
}

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

func BuildAllOverlays(nodes []node.NodeInfo) error {
	for _, n := range nodes {
		var overlays []string

		overlays = append(overlays, n.SystemOverlay.Get())
		overlays = append(overlays, n.RuntimeOverlay.Get())

		wwlog.Printf(wwlog.INFO, "Building overlays for %s: [%s]\n", n.Id.Get(), strings.Join(overlays, ", "))

		for _, overlay := range overlays {
			err := BuildOverlay(n, overlay)
			if err != nil {
				return errors.Wrap(err, "could not build overlay "+n.Id.Get()+"/"+overlay+".img")
			}
		}
	}
	return nil
}

// TODO: Add an Overlay Delete for both sourcedir and image

func BuildSpecificOverlays(nodes []node.NodeInfo, overlayName string) error {
	for _, n := range nodes {

		wwlog.Printf(wwlog.INFO, "Building overlay for %s: %s\n", n.Id.Get(), overlayName)
		err := BuildOverlay(n, overlayName)
		if err != nil {
			return errors.Wrap(err, "could not build overlay "+n.Id.Get()+"/"+overlayName+".img")
		}

	}
	return nil
}

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

func OverlayInit(overlayName string) error {
	path := OverlaySourceDir(overlayName)

	if util.IsDir(path) {
		return errors.New("Overlay already exists: " + overlayName)
	}

	err := os.MkdirAll(path, 0755)

	return err
}

func BuildOverlay(nodeInfo node.NodeInfo, overlayName string) error {
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	nodeDB, _ := node.New()
	allNodes, _ := nodeDB.FindAllNodes()
	var tstruct TemplateStruct
	OverlaySourceDir := OverlaySourceDir(overlayName)
	OverlayImage := OverlayImage(nodeInfo.Id.Get(), overlayName)
	OverlayImageDir := path.Dir(OverlayImage)

	if !util.ValidString(overlayName, "^[a-zA-Z0-9-._:]+$") {
		return errors.New("overlay name contains illegal characters: " + overlayName)
	}

	wwlog.Printf(wwlog.DEBUG, "Checking to see if overlay directory exists: %s\n", OverlaySourceDir)
	if !util.IsDir(OverlaySourceDir) {
		return errors.New("overlay does not exist: " + overlayName)
	}

	err = os.MkdirAll(OverlayImageDir, 0755)
	if err == nil {
		wwlog.Printf(wwlog.DEBUG, "Created parent directory for Overlay Images: %s\n", OverlayImageDir)
	} else {
		return errors.Wrap(err, "could not create overlay image directory")
	}

	tmpDir, err := ioutil.TempDir(os.TempDir(), ".wwctl-overlay-")
	if err == nil {
		wwlog.Printf(wwlog.DEBUG, "Creating temporary directory for overlay files: %s\n", tmpDir)
	} else {
		return errors.Wrap(err, "could not create overlay temporary directory")
	}

	wwlog.Printf(wwlog.VERBOSE, "Processing node/overlay: %s/%s\n", nodeInfo.Id.Get(), overlayName)

	tstruct.Id = nodeInfo.Id.Get()
	tstruct.Hostname = nodeInfo.Id.Get()
	tstruct.Id = nodeInfo.Id.Get()
	tstruct.Hostname = nodeInfo.Id.Get()
	tstruct.ClusterName = nodeInfo.ClusterName.Get()
	tstruct.Container = nodeInfo.ContainerName.Get()
	tstruct.KernelVersion = nodeInfo.KernelOverride.Get()
	tstruct.KernelOverride = nodeInfo.KernelOverride.Get()
	tstruct.KernelArgs = nodeInfo.KernelArgs.Get()
	tstruct.Init = nodeInfo.Init.Get()
	tstruct.Root = nodeInfo.Root.Get()
	tstruct.IpmiIpaddr = nodeInfo.IpmiIpaddr.Get()
	tstruct.IpmiNetmask = nodeInfo.IpmiNetmask.Get()
	tstruct.IpmiPort = nodeInfo.IpmiPort.Get()
	tstruct.IpmiGateway = nodeInfo.IpmiGateway.Get()
	tstruct.IpmiUserName = nodeInfo.IpmiUserName.Get()
	tstruct.IpmiPassword = nodeInfo.IpmiPassword.Get()
	tstruct.IpmiInterface = nodeInfo.IpmiInterface.Get()
	tstruct.RuntimeOverlay = nodeInfo.RuntimeOverlay.Get()
	tstruct.SystemOverlay = nodeInfo.SystemOverlay.Get()
	tstruct.NetDevs = make(map[string]*node.NetDevs)
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

	}
	// Backwards compatibility for templates using "Keys"
	for keyname, key := range nodeInfo.Tags {
		tstruct.Keys[keyname] = key.Get()
	}
	for keyname, key := range nodeInfo.Tags {
		tstruct.Tags[keyname] = key.Get()
	}
	tstruct.AllNodes = allNodes
	for _, export := range controller.Nfs.ExportsExtended {
		if export.Mount {
			var mountOpts string
			if export.MountOptions == "" {
				mountOpts = "defaults"
			} else {
				mountOpts = export.MountOptions
			}
			tstruct.NFSMounts = append(tstruct.NFSMounts, fmt.Sprintf("%s:%s %s nfs %s 0 0\n", controller.Ipaddr, export.Path, export.Path, mountOpts))
		}
	}

	wwlog.Printf(wwlog.DEBUG, "Changing directory to OverlayDir: %s\n", OverlaySourceDir)
	err = os.Chdir(OverlaySourceDir)
	if err != nil {
		return errors.Wrap(err, "could not change directory to overlay dir")
	}

	wwlog.Printf(wwlog.VERBOSE, "Walking the overlay structure: %s\n", OverlaySourceDir)
	err = filepath.Walk(".", func(location string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		wwlog.Printf(wwlog.DEBUG, "Found overlay file: %s\n", location)

		if info.IsDir() {
			wwlog.Printf(wwlog.DEBUG, "Found directory: %s\n", location)

			err = os.MkdirAll(path.Join(tmpDir, location), info.Mode())
			if err != nil {
				return errors.Wrap(err, "could not create directory within overlay")
			}
			err = util.CopyUIDGID(location, path.Join(tmpDir, location))
			if err != nil {
				return errors.Wrap(err, "failed setting permissions on overlay directory")
			}

			wwlog.Printf(wwlog.DEBUG, "Created directory in overlay: %s\n", location)

		} else if filepath.Ext(location) == ".ww" {
			wwlog.Printf(wwlog.VERBOSE, "Evaluating overlay template file: %s\n", location)

			destFile := strings.TrimSuffix(location, ".ww")

			tmpl, err := template.New(path.Base(location)).Option("missingkey=default").Funcs(template.FuncMap{
				// TODO: Fix for missingkey=zero
				"Include":     templateFileInclude,
				"IncludeFrom": templateContainerFileInclude,
				"inc":         func(i int) int { return i + 1 },
				"dec":         func(i int) int { return i - 1 },
				// }).ParseGlob(path.Join(OverlayDir, destFile+".ww*"))
			}).ParseGlob(location)
			if err != nil {
				return errors.Wrap(err, "could not parse template "+location)
			}

			w, err := os.OpenFile(path.Join(tmpDir, destFile), os.O_RDWR|os.O_CREATE, info.Mode())
			if err != nil {
				return errors.Wrap(err, "could not open new file for template")
			}
			defer w.Close()

			err = tmpl.Execute(w, tstruct)
			if err != nil {
				return errors.Wrap(err, "could not execute template")
			}

			err = util.CopyUIDGID(location, path.Join(tmpDir, destFile))
			if err != nil {
				return errors.Wrap(err, "failed setting permissions on template output file")
			}

			wwlog.Printf(wwlog.DEBUG, "Wrote template file into overlay: %s\n", destFile)

			//		} else if b, _ := regexp.MatchString(`\.ww[a-zA-Z0-9\-\._]*$`, location); b {
			//			wwlog.Printf(wwlog.DEBUG, "Ignoring WW template file: %s\n", location)
		} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			wwlog.Printf(wwlog.DEBUG, "Found symlink %s\n", location)
			destination, err := os.Readlink(location)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
			}
			err = os.Symlink(destination, path.Join(tmpDir, location))
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
			}
		} else {

			err := util.CopyFile(location, path.Join(tmpDir, location))
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

	wwlog.Printf(wwlog.DEBUG, "Finished generating overlay working directory for: %s/%s\n", nodeInfo.Id.Get(), overlayName)

	compressor, err := exec.LookPath("pigz")
	if err != nil {
		wwlog.Printf(wwlog.DEBUG, "Could not locate PIGZ, using GZIP\n")
		compressor = "gzip"
	} else {
		wwlog.Printf(wwlog.DEBUG, "Using PIGZ to compress the overlay: %s\n", compressor)
	}

	cmd := fmt.Sprintf("cd \"%s\"; find . | cpio --quiet -o -H newc | %s -c > \"%s\"", tmpDir, compressor, OverlayImage)

	wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
	err = exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		return errors.Wrap(err, "could not generate compressed runtime image overlay")
	}
	wwlog.Printf(wwlog.VERBOSE, "Completed building overlay image: %s\n", OverlayImage)

	wwlog.Printf(wwlog.DEBUG, "Removing temporary directory: %s\n", tmpDir)
	os.RemoveAll(tmpDir)

	return nil
}
