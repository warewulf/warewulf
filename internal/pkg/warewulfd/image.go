package warewulfd

import (
	"net/http"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleImage handles container/image requests
func HandleImage(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	var stageFile string

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
	} else {
		if ctx.remoteNode.ImageName != "" {
			stageFile = image.ImageFile(ctx.remoteNode.ImageName)
		} else {
			wwlog.Warn("No image set for node %s", ctx.remoteNode.Id())
		}
	}

	sendResponse(w, req, stageFile, nil, ctx)
}
