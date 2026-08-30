package warewulfd

import (
	"net/http"
	"path"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleIpxe handles iPXE boot script requests
func HandleIpxe(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	var stageFile string
	var tmplData *templateVars

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
		stageFile = path.Join(ctx.conf.Paths.Sysconfdir, "/warewulf/ipxe/unconfigured.ipxe")
		tmplData = &templateVars{
			Hwaddr: ctx.rinfo.hwaddr}
	} else {
		template := ctx.remoteNode.Ipxe
		if template == "" {
			template = "default"
		}
		stageFile = path.Join(ctx.conf.Paths.Sysconfdir, "warewulf/ipxe", template+".ipxe")
		tmplData = buildTemplateVars(ctx.conf, ctx.rinfo, ctx.remoteNode)
	}

	sendResponse(w, req, stageFile, tmplData, ctx)
}
