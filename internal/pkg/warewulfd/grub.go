package warewulfd

import (
	"net/http"
	"path"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleGrub handles GRUB configuration requests
func HandleGrub(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	stageFile := path.Join(ctx.conf.Paths.Sysconfdir, "warewulf/grub/grub.cfg.ww")
	tmplData := buildTemplateVars(ctx.conf, ctx.rinfo, ctx.remoteNode)
	sendResponse(w, req, stageFile, tmplData, ctx)
}
