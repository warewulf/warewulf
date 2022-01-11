package overlay

import (
	"path"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/internal/pkg/util"
)

func OverlaySourceTopDir() string {
	return path.Join(warewulfconf.DataStore(), "overlays/")
}

func OverlaySourceDir(overlayName string) string {
	return path.Join(OverlaySourceTopDir(), overlayName)
}

func OverlayImage(nodeName string, overlayName string) string {
	return path.Join(warewulfconf.DataStore(), "/provision/overlays/", nodeName, overlayName + ".img")
}
func SystemOverlayDir() string {
	return path.Join(OverlaySourceTopDir(), "/system")
}

func RuntimeOverlayDir() string {
	return path.Join(OverlaySourceTopDir(), "/runtime")
}

func SystemOverlaySource(overlayName string) string {
	if overlayName == "" {
		wwlog.Printf(wwlog.ERROR, "System overlay name is not defined\n")
		return ""
	}

	if !util.ValidString(overlayName, "^[a-zA-Z0-9-._]+$") {
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

	if !util.ValidString(overlayName, "^[a-zA-Z0-9-._]+$") {
		wwlog.Printf(wwlog.ERROR, "Runtime overlay name contains illegal characters: %s\n", overlayName)
		return ""
	}

	return path.Join(RuntimeOverlayDir(), overlayName)
}

func SystemOverlayImage(nodeName string) string {
	if nodeName == "" {
		wwlog.Printf(wwlog.ERROR, "Node name is not defined\n")
		return ""
	}

	if !util.ValidString(nodeName, "^[a-zA-Z0-9-._:]+$") {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", nodeName)
		return ""
	}

	return path.Join(warewulfconf.DataStore(), "/provision/overlays/runtime/", nodeName + ".img")
}

func RuntimeOverlayImage(nodeName string) string {
	if nodeName == "" {
		wwlog.Printf(wwlog.ERROR, "Node name is not defined\n")
		return ""
	}

	if !util.ValidString(nodeName, "^[a-zA-Z0-9-._:]+$") {
		wwlog.Printf(wwlog.ERROR, "System overlay name contains illegal characters: %s\n", nodeName)
		return ""
	}

	return path.Join(warewulfconf.DataStore(), "/provision/overlays/system/", nodeName + ".img")
}
