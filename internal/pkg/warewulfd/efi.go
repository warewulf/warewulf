package warewulfd

import (
	"fmt"
	"net/http"
	"path"

	"github.com/warewulf/warewulf/internal/pkg/image"
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
			wwlog.Error("could't find grub*.efi for %s", imageName)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case "grub.cfg":
		stageFile = path.Join(ctx.conf.Paths.Sysconfdir, "warewulf/grub/grub.cfg.ww")
		tmplData = buildTemplateVars(ctx.conf, ctx.rinfo, ctx.remoteNode)
		if stageFile == "" {
			wwlog.Error("could't find grub.cfg template for %s", imageName)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		wwlog.ErrorExc(fmt.Errorf("could't find efiboot file: %s", ctx.rinfo.efifile), "")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sendResponse(w, req, stageFile, tmplData, ctx)
}
