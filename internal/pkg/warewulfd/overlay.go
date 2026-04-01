package warewulfd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleSystemOverlay handles system overlay requests.
func HandleSystemOverlay(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
		sendResponse(w, req, "", nil, ctx)
		return
	}

	stageFile, err := getOverlayFile(
		ctx.remoteNode,
		"system",
		ctx.conf.Warewulf.AutobuildOverlays())

	if err != nil {
		if errors.Is(err, overlay.ErrDoesNotExist) {
			w.WriteHeader(http.StatusNotFound)
			wwlog.ErrorExc(err, "")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		wwlog.ErrorExc(err, "")
		return
	}

	sendResponse(w, req, stageFile, nil, ctx)
}

// HandleRuntimeOverlay handles runtime overlay requests.
// If TLS is enabled, returns 403 Forbidden for plain-HTTP requests.
func HandleRuntimeOverlay(w http.ResponseWriter, req *http.Request) {
	if config.Get().Warewulf.TLSEnabled() && req.TLS == nil {
		wwlog.Denied("runtime overlay requested over insecure connection")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
		sendResponse(w, req, "", nil, ctx)
		return
	}

	stageFile, err := getOverlayFile(
		ctx.remoteNode,
		"runtime",
		ctx.conf.Warewulf.AutobuildOverlays())

	if err != nil {
		if errors.Is(err, overlay.ErrDoesNotExist) {
			w.WriteHeader(http.StatusNotFound)
			wwlog.ErrorExc(err, "")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		wwlog.ErrorExc(err, "")
		return
	}

	sendResponse(w, req, stageFile, nil, ctx)
}

// HandleOverlayFile serves an individual file from a named overlay's rootfs.
// The URL structure is /overlay-file/{overlay}/{path}.
// Every request must identify a node via ?wwid= or ARP fallback.
// If ?render is present, the file is rendered as a Go template for the
// identified node. If the path does not end in .ww but a .ww-suffixed version
// exists, that file is used. Otherwise the raw file is returned.
func HandleOverlayFile(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "overlay and path required", http.StatusBadRequest)
		return
	}
	overlayName := parts[2]
	filePath := strings.Join(parts[3:], "/")
	if overlayName == "" {
		http.Error(w, "no overlay specified", http.StatusBadRequest)
		return
	}
	if filePath == "" {
		http.Error(w, "no path specified", http.StatusBadRequest)
		return
	}

	remoteNode, ok := authenticateNode(w, req)
	if !ok {
		return
	}

	myOverlay, err := overlay.Get(overlayName)
	if err != nil {
		wwlog.Error("overlay-file: overlay not found: %s", overlayName)
		http.Error(w, fmt.Sprintf("overlay not found: %s", overlayName), http.StatusNotFound)
		return
	}

	overlayFile := myOverlay.File(filePath)
	if !path.IsAbs(overlayFile) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(overlayFile); os.IsNotExist(err) {
		_, hasRender := req.URL.Query()["render"]
		wwFile := myOverlay.File(filePath + ".ww")
		if hasRender && path.IsAbs(wwFile) {
			if _, err := os.Stat(wwFile); err == nil {
				wwlog.Debug("overlay-file: using .ww suffix for %s", filePath)
				overlayFile = wwFile
			} else {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}

	_, render := req.URL.Query()["render"]
	if render {
		if renderName := req.URL.Query().Get("render"); renderName != "" && renderName != remoteNode.Id() {
			http.Error(w, fmt.Sprintf("render node %q does not match identified node %q", renderName, remoteNode.Id()), http.StatusBadRequest)
			return
		}
		if !strings.HasSuffix(overlayFile, ".ww") {
			http.Error(w, "render requires a .ww template file", http.StatusBadRequest)
			return
		}

		registry, err := node.New()
		if err != nil {
			wwlog.Error("overlay-file: error opening node database: %s", err)
			http.Error(w, fmt.Sprintf("error opening node database: %s", err), http.StatusInternalServerError)
			return
		}

		allNodes, err := registry.FindAllNodes()
		if err != nil {
			wwlog.Error("overlay-file: error loading nodes: %s", err)
			http.Error(w, fmt.Sprintf("error loading nodes: %s", err), http.StatusInternalServerError)
			return
		}

		tstruct, err := overlay.InitStruct(overlayName, remoteNode, allNodes)
		if err != nil {
			wwlog.Error("overlay-file: error initializing template data: %s", err)
			http.Error(w, fmt.Sprintf("error initializing template data: %s", err), http.StatusInternalServerError)
			return
		}
		tstruct.BuildSource = overlayFile

		buffer, _, _, err := overlay.RenderTemplateFile(overlayFile, tstruct)
		if err != nil {
			wwlog.Error("overlay-file: error rendering template %s: %s", overlayFile, err)
			http.Error(w, fmt.Sprintf("error rendering template: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
		if _, err := buffer.WriteTo(w); err != nil {
			wwlog.Error("overlay-file: error writing response: %s", err)
		}
		wwlog.Info("overlay-file: rendered %s for node %s", overlayFile, remoteNode.Id())
	} else {
		fileBytes, err := os.ReadFile(overlayFile)
		if err != nil {
			wwlog.Error("overlay-file: error reading file: %s", err)
			http.Error(w, fmt.Sprintf("error reading file: %s", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(fileBytes)))
		if _, err := w.Write(fileBytes); err != nil {
			wwlog.Error("overlay-file: error writing response: %s", err)
		}
		wwlog.Info("overlay-file: sent %s for node %s", overlayFile, remoteNode.Id())
	}
}
