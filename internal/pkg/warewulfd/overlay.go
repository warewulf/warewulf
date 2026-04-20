package warewulfd

import (
	"errors"
	"net/http"

	"github.com/warewulf/warewulf/internal/pkg/config"
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
