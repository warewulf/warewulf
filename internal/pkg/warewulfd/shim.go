package warewulfd

import (
	"net/http"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleShim handles direct shim binary requests
func HandleShim(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	var stageFile string

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
	} else {
		if ctx.remoteNode.ImageName != "" {
			stageFile = image.ShimFind(ctx.remoteNode.ImageName)

			if stageFile == "" {
				wwlog.Error("No kernel found for image %s", ctx.remoteNode.ImageName)
			}
		} else {
			wwlog.Warn("No image set for this %s", ctx.remoteNode.Id())
		}
	}

	sendResponse(w, req, stageFile, nil, ctx)
}
