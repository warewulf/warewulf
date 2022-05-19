package warewulfd

import (
	"net/http"
	"path"
	"strconv"
	"bytes"
	"text/template"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	nodepkg "github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

type iPxeTemplate struct {
	Message        string
	WaitTime       string
	Hostname       string
	Fqdn           string
	Id             string
	Cluster        string
	ContainerName  string
	Hwaddr         string
	Ipaddr         string
	Port           string
	KernelArgs     string
	KernelOverride string
}

func ProvisionSend(w http.ResponseWriter, req *http.Request) {
	conf, err := warewulfconf.New()
	if err != nil {
		wwlog.Error("Could not open Warewulf configuration: %s", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	rinfo, err := parseReq(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wwlog.ErrorExc(err, "")
		return
	}

	wwlog.Recv("hwaddr: %s, ipaddr: %s, stage: %s", rinfo.hwaddr, req.RemoteAddr, rinfo.stage )

	if rinfo.stage == "runtime" && conf.Warewulf.Secure {
		if rinfo.remoteport >= 1024 {
			wwlog.Denied("Non-privledged port: %s", req.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	status_stages := map[string]string{
		"ipxe": "IPXE",
		"kernel": "KERNEL",
		"kmods": "KMODS_OVERLAY",
		"system": "SYSTEM_OVERLAY",
		"runtime": "RUNTIME_OVERLAY" }

	status_stage := status_stages[rinfo.stage]
	var stage_overlays []string
	var stage_file string = ""
	// TODO: when module version is upgraded to go1.18, should be 'any' type
	var tmpl_data interface{}

	node, err := GetNode(rinfo.hwaddr)
	if err != nil {
		// If we failed to find a node, let's see if we can add one...
		var netdev string

		wwlog.Warn("%s (node not configured)", rinfo.hwaddr)

		nodeDB, err := nodepkg.New()
		if err != nil {
			wwlog.Error("Could not read node configuration file: %s", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		n, netdev, err := nodeDB.FindDiscoverableNode()
		if err == nil {
			n.NetDevs[netdev].Hwaddr.Set(rinfo.hwaddr)
			n.Discoverable.SetB(false)
			err := nodeDB.NodeUpdate(n)
			if err != nil {
				wwlog.Serv("%s (failed to set node configuration)", rinfo.hwaddr)

			} else {
				err := nodeDB.Persist()
				if err != nil {
					wwlog.Serv("%s (failed to persist node configuration)", rinfo.hwaddr)

				} else {
					node = n
					_ = overlay.BuildAllOverlays([]nodepkg.NodeInfo{n})

					wwlog.Serv("%s (node automatically configured)", rinfo.hwaddr)

					err := LoadNodeDB()
					if err != nil {
						wwlog.Warn("Could not reload configuration: %s", err)
					}

				}
			}
		}
	}

	if node.AssetKey.Defined() && node.AssetKey.Get() != rinfo.assetkey {
		w.WriteHeader(http.StatusUnauthorized)
		wwlog.Denied("Incorrect asset key for node: %s", node.Id.Get())
		updateStatus(node.Id.Get(), status_stage, "BAD_ASSET", rinfo.ipaddr)
		return
	}

	if !node.Id.Defined() {
		wwlog.Error("%s (unknown/unconfigured node)", rinfo.hwaddr)
		if rinfo.stage == "ipxe" {
			stage_file = path.Join(buildconfig.SYSCONFDIR(), "/warewulf/ipxe/unconfigured.ipxe")
			tmpl_data = iPxeTemplate{
				Hwaddr : rinfo.hwaddr }
		}

	}else if rinfo.stage == "ipxe" {
		stage_file = path.Join(buildconfig.SYSCONFDIR(), "warewulf/ipxe/"+node.Ipxe.Get()+".ipxe")
		tmpl_data = iPxeTemplate{
			Id : node.Id.Get(),
			Cluster : node.ClusterName.Get(),
			Fqdn : node.Id.Get(),
			Ipaddr : conf.Ipaddr,
			Port : strconv.Itoa(conf.Warewulf.Port),
			Hostname : node.Id.Get(),
			Hwaddr : rinfo.hwaddr,
			ContainerName : node.ContainerName.Get(),
			KernelArgs : node.Kernel.Args.Get(),
			KernelOverride : node.Kernel.Override.Get() }

	}else if rinfo.stage == "kernel" {
		if node.Kernel.Override.Defined() {
			stage_file = kernel.KernelImage(node.Kernel.Override.Get())
		} else if node.ContainerName.Defined() {
			stage_file = container.KernelFind(node.ContainerName.Get())

			if stage_file == "" {
				wwlog.Error("No kernel found for container %s", node.ContainerName.Get())
			}
		} else {
			wwlog.Warn("No kernel version set for node %s", node.Id.Get())
		}

	}else if rinfo.stage == "kmods" {
		if node.Kernel.Override.Defined() {
			stage_file = kernel.KmodsImage(node.Kernel.Override.Get())
		}else{
			wwlog.Warn("No kernel override modules set for node %s", node.Id.Get())
		}

	}else if rinfo.stage == "container" {
		if node.ContainerName.Defined() {
			stage_file = container.ImageFile(node.ContainerName.Get())
		} else {
			wwlog.Warn("No container set for node %s", node.Id.Get())
		}

	}else if rinfo.stage == "system" {
		if len(node.SystemOverlay.GetSlice()) != 0 {
			stage_overlays = node.SystemOverlay.GetSlice()
		} else {
			wwlog.Warn("No system overlay set for node %s", node.Id.Get())
		}

	}else if rinfo.stage == "runtime" {
		if rinfo.overlay != "" {
			stage_overlays = []string{rinfo.overlay}
		} else if len(node.RuntimeOverlay.GetSlice()) != 0 {
			stage_overlays = node.RuntimeOverlay.GetSlice()
		} else {
			wwlog.Warn("No runtime overlay set for node %s", node.Id.Get())
		}

	}

	if len(stage_overlays) > 0 {
		stage_file = overlay.OverlayImage(node.Id.Get(), stage_overlays)
		if conf.Warewulf.AutobuildOverlays {
			oneoverlaynewer := false
			for _, overlayname := range stage_overlays {
				oneoverlaynewer = oneoverlaynewer || util.PathIsNewer(stage_file, overlay.OverlaySourceDir(overlayname))
			}
			if !util.IsFile(stage_file) || util.PathIsNewer(stage_file, nodepkg.ConfigFile) || oneoverlaynewer {
				wwlog.Serv("BUILD %15s, overlays %v", node.Id.Get(), stage_overlays)
				_ = overlay.BuildOverlay(node, stage_overlays)
			}
		}
	}

	wwlog.Serv("stage_file '%s'", stage_file )

	if util.IsFile(stage_file) {

		if tmpl_data != nil {
			if rinfo.compress != "" {
				wwlog.Error("Unsupported %s compressed version for file: %s",
					rinfo.compress, stage_file)
				w.WriteHeader(http.StatusNotFound)
				return
			}

			tmpl, err := template.ParseFiles(stage_file)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				wwlog.ErrorExc(err, "")
				return
			}

			// template engine writes file to buffer in case rendering fails
			var buf bytes.Buffer

			err = tmpl.Execute(&buf, tmpl_data)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				wwlog.ErrorExc(err, "")
				return
			}

			w.Header().Set("Content-Type", "text")
			w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
			_, err = buf.WriteTo(w)
			if err != nil {
				wwlog.ErrorExc(err, "")
			}

			wwlog.Send("%15s: %s", node.Id.Get(), stage_file)

		}else{
			if rinfo.compress == "gz" {
				stage_file += ".gz"

				if !util.IsFile(stage_file) {
					wwlog.Error("unprepared for compressed version of file %s",
						stage_file)
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}else if rinfo.compress != "" {
				wwlog.Error("unsupported %s compressed version of file %s",
					rinfo.compress, stage_file)
				w.WriteHeader(http.StatusNotFound)
			}

			err = sendFile(w, stage_file, node.Id.Get())
			if err != nil {
				wwlog.ErrorExc(err, "")
				return
			}
		}

		updateStatus(node.Id.Get(), status_stage, path.Base(stage_file), rinfo.ipaddr)

	}else if stage_file == "" {
		w.WriteHeader(http.StatusBadRequest)
		wwlog.Error("No resource selected")
		updateStatus(node.Id.Get(), status_stage, "BAD_REQUEST", rinfo.ipaddr)

	}else{
		w.WriteHeader(http.StatusNotFound)
		wwlog.Error("Not found: %s", stage_file )
		updateStatus(node.Id.Get(), status_stage, "NOT_FOUND", rinfo.ipaddr)
	}

}
