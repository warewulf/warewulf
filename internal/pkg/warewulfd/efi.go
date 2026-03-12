package warewulfd

import (
	"fmt"
	"net/http"
	"path"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleEfiBoot handles EFI boot file requests (shim, grub, grub.cfg)
func HandleEfiBoot(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
		sendResponse(w, req, "", nil, ctx)
		return
	}

	wwlog.Debug("requested method: %s", req.Method)
	imageName := ctx.remoteNode.ImageName
	var stageFile string
	var tmplData *templateVars

	switch ctx.rinfo.efifile {
	case "shim.efi":
		stageFile = image.ShimFind(imageName)
		if stageFile == "" {
			wwlog.Error("couldn't find shim.efi for %s", imageName)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case "grub.efi", "grub-tpm.efi", "grubx64.efi", "grubia32.efi", "grubaa64.efi", "grubarm.efi":
		stageFile = image.GrubFind(imageName)
		if stageFile == "" {
			wwlog.Error("couldn't find grub*.efi for %s", imageName)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case "grub.cfg":
		// this is the first stage fot the node in the boot process where the
		// host is involved, so we reset the SentLog and *prepend* is with the
		// grub.cfg which was sent out via tftp, so clear the logs and add as
		// first entry the grub.cfg we sent out via tftp
		ctx.tpm.ClearLogs()
		if err = ctx.tpm.Update(path.Join(ctx.conf.TFTP.TftpRoot, "warewulf/grub.cfg"), ""); err != nil {
			wwlog.Warn("couldn't update TPM log with grub.cfg sent by tftp", err)
		}
		stageFile = path.Join(ctx.conf.Paths.Sysconfdir, "warewulf/grub/grub.cfg.ww")
		tmplData = buildTemplateVars(ctx.conf, ctx.rinfo, ctx.remoteNode)
		if !util.IsFile(stageFile) {
			wwlog.Error("couldn't find grub.cfg template for %s", imageName)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		wwlog.ErrorExc(fmt.Errorf("couldn't find efiboot file: %s", ctx.rinfo.efifile), "")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sendResponse(w, req, stageFile, tmplData, ctx)
}
