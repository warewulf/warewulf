package config

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"path"
)

const (
	LocalStateDir = "/var/warewulf"
)

func NodeConfig() string {
	return fmt.Sprintf("%s/nodes.conf", LocalStateDir)
}

func OverlayDir() string {
	return fmt.Sprintf("%s/overlays/", LocalStateDir)
}

func SystemOverlayDir() string {
	return path.Join(OverlayDir(), "/system")
}

func RuntimeOverlayDir() string {
	return path.Join(OverlayDir(), "/runtime")
}

func VnfsImageParentDir() string {
	return fmt.Sprintf("%s/provision/vnfs/", LocalStateDir)
}

func VnfsChrootParentDir() string {
	return fmt.Sprintf("%s/chroot/", LocalStateDir)
}

func KernelParentDir() string {
	return fmt.Sprintf("%s/provision/kernel/", LocalStateDir)
}

func SystemOverlaySource(overlayName string) string {
	if overlayName == "" {
		wwlog.Printf(wwlog.ERROR, "System overlay name is not defined\n")
		return ""
	}

	if util.TaintCheck(overlayName, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", overlayName)
		return ""
	}

	return path.Join(SystemOverlayDir(), overlayName)
}

func RuntimeOverlaySource(overlayName string) string {
	if overlayName == "" {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name is not defined\n")
		return ""
	}

	if util.TaintCheck(overlayName, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", overlayName)
		return ""
	}

	return path.Join(RuntimeOverlayDir(), overlayName)
}

func KernelImage(kernelVersion string) string {
	if kernelVersion == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Version is not defined\n")
		return ""
	}

	if util.TaintCheck(kernelVersion, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelVersion)
		return ""
	}

	return path.Join(KernelParentDir(), kernelVersion, "vmlinuz")
}

func KmodsImage(kernelVersion string) string {
	if kernelVersion == "" {
		wwlog.Printf(wwlog.ERROR, "Kernel Version is not defined\n")
		return ""
	}

	if util.TaintCheck(kernelVersion, "^[a-zA-Z0-9-._]+$") == false {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", kernelVersion)
		return ""
	}

	return path.Join(KernelParentDir(), kernelVersion, "kmods.img")
}

func SystemOverlayImage(nodeName string) string {
	if nodeName == "" {
		wwlog.Printf(wwlog.ERROR, "Node name is not defined\n")
		return ""
	}

	if util.TaintCheck(nodeName, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", nodeName)
		return ""
	}

	return fmt.Sprintf("%s/provision/overlays/system/%s.img", LocalStateDir, nodeName)
}

func RuntimeOverlayImage(nodeName string) string {
	if nodeName == "" {
		wwlog.Printf(wwlog.ERROR, "Node name is not defined\n")
		return ""
	}

	if util.TaintCheck(nodeName, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", nodeName)
		return ""
	}

	return fmt.Sprintf("%s/provision/overlays/runtime/%s.img", LocalStateDir, nodeName)
}

func VnfsImageDir(uri string) string {
	if uri == "" {
		wwlog.Printf(wwlog.ERROR, "VNFS URI is not defined\n")
		return ""
	}

	if util.TaintCheck(uri, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "VNFS name contains illegal characters: %s\n", uri)
		return ""
	}

	return path.Join(VnfsImageParentDir(), uri)
}

func VnfsImage(uri string) string {
	return path.Join(VnfsImageDir(uri), "image")
}

func VnfsChroot(uri string) string {
	if uri == "" {
		wwlog.Printf(wwlog.ERROR, "VNFS name is not defined\n")
		return ""
	}

	if util.TaintCheck(uri, "^[a-zA-Z0-9-._:]+$") == false {
		wwlog.Printf(wwlog.ERROR, "VNFS name contains illegal characters: %s\n", uri)
		return ""
	}

	return path.Join(VnfsChrootParentDir(), uri)
}
