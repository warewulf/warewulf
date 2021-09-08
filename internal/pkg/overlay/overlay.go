package overlay

import (
	"fmt"
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

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

type TemplateStruct struct {
	Id            string
	Hostname      string
	ClusterName   string
	Container     string
	Init          string
	Root          string
	IpmiIpaddr    string
	IpmiNetmask   string
	IpmiPort      string
	IpmiGateway   string
	IpmiUserName  string
	IpmiPassword  string
	IpmiInterface string
	NetDevs       map[string]*node.NetDevs
	Keys          map[string]string
	AllNodes      []node.NodeInfo
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
	return findAllOverlays("runtime")
}

func SystemOverlayInit(name string) error {
	return overlayInit(name, "system")
}

func RuntimeOverlayInit(name string) error {
	return overlayInit(name, "runtime")
}

func findAllOverlays(overlayType string) ([]string, error) {
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
		if file.IsDir() {
			ret = append(ret, file.Name())
		}
	}

	return ret, nil
}

func overlayInit(name string, overlayType string) error {
	var path string

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

	if util.IsDir(path) {
		return errors.New("Overlay already exists: " + name)
	}

	err := os.MkdirAll(path, 0755)

	return err
}

func buildOverlay(nodeList []node.NodeInfo, overlayType string) error {
	nodeDB, _ := node.New()
	allNodes, _ := nodeDB.FindAllNodes()

	for _, n := range nodeList {
		var t TemplateStruct
		var OverlayDir string
		var OverlayFile string

		if overlayType == "runtime" {
			wwlog.Printf(wwlog.VERBOSE, "Building runtime overlay for: %s\n", n.Id.Get())

			OverlayDir = config.RuntimeOverlaySource(n.RuntimeOverlay.Get())
			OverlayFile = config.RuntimeOverlayImage(n.Id.Get())
		} else if overlayType == "system" {
			wwlog.Printf(wwlog.VERBOSE, "Building system overlay for: %s\n", n.Id.Get())

			OverlayDir = config.SystemOverlaySource(n.SystemOverlay.Get())
			OverlayFile = config.SystemOverlayImage(n.Id.Get())
		} else {
			wwlog.Printf(wwlog.ERROR, "overlayType requested is not supported: %s\n", overlayType)
			os.Exit(1)
		}

		wwlog.Printf(wwlog.DEBUG, "Processing overlay for node: %s\n", n.Id.Get())

		t.Id = n.Id.Get()
		t.Hostname = n.Id.Get()
		t.ClusterName = n.ClusterName.Get()
		t.Container = n.ContainerName.Get()
		t.Init = n.Init.Get()
		t.Root = n.Root.Get()
		t.IpmiIpaddr = n.IpmiIpaddr.Get()
		t.IpmiNetmask = n.IpmiNetmask.Get()
		t.IpmiPort = n.IpmiPort.Get()
		t.IpmiGateway = n.IpmiGateway.Get()
		t.IpmiUserName = n.IpmiUserName.Get()
		t.IpmiPassword = n.IpmiPassword.Get()
		t.IpmiInterface = n.IpmiInterface.Get()
		t.NetDevs = make(map[string]*node.NetDevs)
		t.Keys = make(map[string]string)
		for devname, netdev := range n.NetDevs {
			var nd node.NetDevs
			t.NetDevs[devname] = &nd
			t.NetDevs[devname].Hwaddr = netdev.Hwaddr.Get()
			t.NetDevs[devname].Ipaddr = netdev.Ipaddr.Get()
			t.NetDevs[devname].Netmask = netdev.Netmask.Get()
			t.NetDevs[devname].Gateway = netdev.Gateway.Get()
			t.NetDevs[devname].Type = netdev.Type.Get()
			t.NetDevs[devname].Default = netdev.Default.GetB()
			mask := net.IPMask(net.ParseIP(netdev.Netmask.Get()).To4())
			ipaddr := net.ParseIP(netdev.Ipaddr.Get()).To4()
			netaddr := net.IPNet{IP: ipaddr, Mask: mask}
			netPrefix, _ := net.IPMask(net.ParseIP(netdev.Netmask.Get()).To4()).Size()
			t.NetDevs[devname].Prefix = strconv.Itoa(netPrefix)
			t.NetDevs[devname].IpCIDR = netaddr.String()

		}
		for keyname, key := range n.Keys {
			t.Keys[keyname] = key.Get()
		}
		t.AllNodes = allNodes

		if overlayType == "runtime" && !n.RuntimeOverlay.Defined() {
			wwlog.Printf(wwlog.WARN, "Undefined runtime overlay, skipping node: %s\n", n.Id.Get())
		}
		if overlayType == "system" && !n.SystemOverlay.Defined() {
			wwlog.Printf(wwlog.WARN, "Undefined system overlay, skipping node: %s\n", n.Id.Get())
		}

		wwlog.Printf(wwlog.DEBUG, "Checking to see if overlay directory exists: %s\n", OverlayDir)
		if !util.IsDir(OverlayDir) {
			wwlog.Printf(wwlog.WARN, "%-35s: Skipped (runtime overlay template not found)\n", n.Id.Get())
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

			wwlog.Printf(wwlog.DEBUG, "Overlay Walk for '%s': OVERLAY:/%s\n", n.Id.Get(), location)

			if info.IsDir() {
				wwlog.Printf(wwlog.DEBUG, "Found directory: %s\n", location)

				err = os.MkdirAll(path.Join(tmpDir, location), info.Mode())
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					return err
				}
				err = util.CopyUIDGID(location, path.Join(tmpDir, location))
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "%s\n", err)
					return err
				}

			} else if filepath.Ext(location) == ".ww" {
				wwlog.Printf(wwlog.DEBUG, "Found template file: %s\n", location)

				destFile := strings.TrimSuffix(location, ".ww")

				tmpl, err := template.New(path.Base(location)).Funcs(template.FuncMap{
					"Include":     templateFileInclude,
					"IncludeFrom": templateContainerFileInclude,
					"inc":         func(i int) int { return i + 1 },
					"dec":         func(i int) int { return i - 1 },
				}).ParseGlob(path.Join(OverlayDir, destFile+".ww*"))
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "template.New %s\n", err)
					return nil
				}

				w, err := os.OpenFile(path.Join(tmpDir, destFile), os.O_RDWR|os.O_CREATE, info.Mode())
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "path.Join %s\n", err)
					return err
				}
				defer w.Close()

				wwlog.Printf(wwlog.VERBOSE, "Writing overlay template: OVERLAY:/%s\n", destFile)
				err = tmpl.Execute(w, t)
				if err != nil {
					wwlog.Printf(wwlog.ERROR, "tmpl.Execute %s\n", err)
					return nil
				}

				err = util.CopyUIDGID(location, path.Join(tmpDir, destFile))
				if err != nil {
					return err
				}

			} else if b, _ := regexp.MatchString(`\.ww[a-zA-Z0-9\-\._]*$`, location); b {
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
		if err != nil {
			return errors.Wrap(err, "failed to open dir")
		}

		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Error with filepath walk: %s\n", err)
			os.Exit(1)
		}

		wwlog.Printf(wwlog.VERBOSE, "Finished generating overlay directory for: %s\n", n.Id.Get())

		compressor, err := exec.LookPath("pigz")
		if err != nil {
			wwlog.Printf(wwlog.VERBOSE, "Could not locate PIGZ, using GZIP\n")
			compressor = "gzip"
		} else {
			wwlog.Printf(wwlog.VERBOSE, "Using PIGZ to compress the container: %s\n", compressor)
		}

		cmd := fmt.Sprintf("cd \"%s\"; find . | cpio --quiet -o -H newc | %s -c > \"%s\"", tmpDir, compressor, OverlayFile)

		wwlog.Printf(wwlog.DEBUG, "RUNNING: %s\n", cmd)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not generate runtime image overlay: %s\n", err)
			continue
		}
		wwlog.Printf(wwlog.INFO, "%-35s: Done\n", n.Id.Get())

		wwlog.Printf(wwlog.DEBUG, "Removing temporary directory: %s\n", tmpDir)
		os.RemoveAll(tmpDir)

	}

	return nil
}
