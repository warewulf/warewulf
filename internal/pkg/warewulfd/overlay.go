package warewulfd

import (
	"net/http"
	"os"
	"path"
	"strconv"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func OverlaySend(w http.ResponseWriter, req *http.Request) {
	conf := warewulfconf.Get()
	overlaySourceDir := overlay.OverlaySourceDir("wwwroot")
	if !util.IsDir(overlaySourceDir) {
		wwlog.Error("Overlay source dir wwwroot doesn't exist")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	rinfo, err := parseReqRender(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wwlog.ErrorExc(err, "")
		return
	}

	wwlog.Recv("render req node: %s, overlay: %s", rinfo.node, rinfo.overlay)
	if conf.Warewulf.Secure {
		if rinfo.remoteport >= 1024 {
			wwlog.Denied("Non-privileged port: %s", req.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	overlayFile := path.Join(overlaySourceDir, rinfo.overlay)
	if !path.IsAbs(overlayFile) {
		wwlog.Denied("Path %s isn't absolute", overlayFile)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var node node.NodeInfo
	nodeName := rinfo.node
	if nodeName != "" {
		node, err = GetNodeById(nodeName)
		if err != nil {
			wwlog.Warn("Unknown node %s from: %s", nodeName, req.RemoteAddr)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if !util.IsFile(overlayFile) {
			overlayFile += ".ww"
			wwlog.Debug("trying to use .ww for file: %s", overlayFile)
		}
		if !util.IsFile(overlayFile) {
			wwlog.Denied("file doesn't exists: %s", overlayFile)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		tstruct := overlay.InitStruct(&node)
		tstruct.BuildSource = overlayFile
		buffer, _, _, err := overlay.RenderTemplateFile(overlayFile, tstruct)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			wwlog.ErrorExc(err, "")
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
		_, err = buffer.WriteTo(w)
		if err != nil {
			wwlog.ErrorExc(err, "")
		}
		wwlog.Send("%15s: %s", node.Id.Get(), overlayFile)
	} else {
		if !util.IsFile(overlayFile) {
			wwlog.Denied("file doesn't exists: %s", overlayFile)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fileBytes, err := os.ReadFile(overlayFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			wwlog.ErrorExc(err, "")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err = w.Write(fileBytes)
		if err != nil {
			wwlog.ErrorExc(err, "")
		}
		wwlog.Send("overlay for node %s: %s", nodeName, overlayFile)

	}
}
