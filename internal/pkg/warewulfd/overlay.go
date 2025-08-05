package warewulfd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func OverlaySend(w http.ResponseWriter, req *http.Request) {
	rinfo, err := parseReqRender(req)
	if err != nil {
		message := "error parsing request: %s"
		wwlog.ErrorExc(err, message, err)
		http.Error(w, fmt.Sprintf(message, err), http.StatusBadRequest)
		return
	}

	myOverlay, err := overlay.GetOverlay(rinfo.overlay)
	if err != nil {
		message := "overlay not found: %s"
		wwlog.Error(message, rinfo.overlay)
		http.Error(w, fmt.Sprintf(message, rinfo.overlay), http.StatusNoContent)
		return
	}

	wwlog.Info("recv: render req overlay: %s, path: %s, node: %s", rinfo.overlay, rinfo.path, rinfo.node)
	if config.Get().Warewulf.Secure() && rinfo.remoteport >= 1024 {
		message := "non-privileged port: %s"
		wwlog.Denied(message, req.RemoteAddr)
		http.Error(w, fmt.Sprintf(message, req.RemoteAddr), http.StatusUnauthorized)
		return
	}

	overlayFile := myOverlay.File(rinfo.path)
	if !path.IsAbs(overlayFile) {
		message := "Path %s isn't absolute"
		wwlog.Denied(message, overlayFile)
		http.Error(w, fmt.Sprintf(message, overlayFile), http.StatusNotFound)
		return
	}

	if !util.IsFile(overlayFile) {
		if rinfo.node != "" && util.IsFile(overlayFile+".ww") {
			wwlog.Debug("appending .ww for file: %s", overlayFile)
			overlayFile += ".ww"
		} else {
			message := "file doesn't exists: %s"
			wwlog.Denied(message, overlayFile)
			http.Error(w, fmt.Sprintf(message, overlayFile), http.StatusNotFound)
			return
		}
	}

	if strings.HasSuffix(overlayFile, ".ww") && rinfo.node != "" {
		nodeDB, err := node.New()
		if err != nil {
			message := "error opening node database: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusNotFound)
			return
		}

		node, err := nodeDB.GetNode(rinfo.node)
		if err != nil {
			message := "error getting node: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusNotFound)
			return
		}

		allNodes, err := nodeDB.FindAllNodes()
		if err != nil {
			message := "error loading nodes from registry: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusInternalServerError)
			return
		}

		tstruct, err := overlay.InitStruct(overlayFile, node, allNodes)
		if err != nil {
			message := "error initializing template data: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusInternalServerError)
			return
		}
		tstruct.BuildSource = overlayFile

		buffer, _, _, err := overlay.RenderTemplateFile(overlayFile, tstruct)
		if err != nil {
			message := "error rendering overlay template: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
		_, err = buffer.WriteTo(w)
		if err != nil {
			message := "error writing overlay template over http connection: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusInternalServerError)
		}
		wwlog.Info("%s: %s", node.Id(), overlayFile)
	} else {
		fileBytes, err := os.ReadFile(overlayFile)
		if err != nil {
			message := "error reading file: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(fileBytes)
		if err != nil {
			message := "error writing overlay file over http connection: %s"
			wwlog.ErrorExc(err, message, err)
			http.Error(w, fmt.Sprintf(message, err), http.StatusInternalServerError)
		}
		wwlog.Info("send overlay file for node %s: %s", rinfo.node, overlayFile)
	}
}

type parserInfoRender struct {
	overlay    string
	path       string
	node       string
	remoteport int
}

func parseReqRender(req *http.Request) (ret parserInfoRender, err error) {
	parts := strings.Split(req.URL.Path, "/")
	ret.overlay = parts[2]
	if ret.overlay == "" {
		return ret, fmt.Errorf("no overlay specified")
	}
	ret.path = strings.Join(parts[3:], "/")
	if ret.path == "" {
		return ret, fmt.Errorf("no path specified")
	}
	if len(req.URL.Query()["render"]) > 0 {
		ret.node = req.URL.Query()["render"][0]
	}
	if _, remoteport, err := net.SplitHostPort(req.RemoteAddr); err != nil {
		return ret, fmt.Errorf("could not obtain remote port from HTTP request: %w", err)
	} else if ret.remoteport, err = strconv.Atoi(remoteport); err != nil {
		return ret, fmt.Errorf("couldn't obtain remote port from HTTP request: %w", err)
	}
	return ret, nil
}
