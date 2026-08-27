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
	// this is the first stage fot the node in the boot process where the
	// host is involved, so we reset the SentLog and *prepend* is with the
	// grub.cfg which was sent out via tftp, so clear the logs and add as
	// first entry the grub.cfg we sent out via tftp
	ctx.tpm.ClearLogs()
	if err = ctx.tpm.Update(path.Join(ctx.conf.TFTP.TftpRoot, "warewulf/grub.cfg"), ""); err != nil {
		wwlog.Warn("couldn't update TPM log with grub.cfg sent by tftp", err)
	}

	stageFile := path.Join(ctx.conf.Paths.Sysconfdir, "warewulf/grub/grub.cfg.ww")
	tmplData := buildTemplateVars(ctx.conf, ctx.rinfo, ctx.remoteNode)
	sendResponse(w, req, stageFile, tmplData, ctx)
}
