package warewulfd

import (
	"net/http"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleGrub handles direct GRUB binary requests
func HandleGrub(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	var stageFile string

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
	} else {
		if ctx.remoteNode.ImageName != "" {
			stageFile = image.GrubFind(ctx.remoteNode.ImageName)
			if stageFile == "" {
				wwlog.Error("No grub found for image %s", ctx.remoteNode.ImageName)
			}
		} else {
			wwlog.Warn("No conainer set for node %s", ctx.remoteNode.Id())
		}
	}

	sendResponse(w, req, stageFile, nil, ctx)
}
