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
		sendResponse(w, req, "", "", nil, ctx)
		return
	}

	wwlog.Debug("requested method: %s", req.Method)
	imageName := ctx.remoteNode.ImageName
	// without an image, the bootloaders are found on the server itself
	imageDir := "/"
	if imageName != "" {
		imageDir = imageRootFs(imageName)
	}
	var dir, stageFile string
	var tmplData *templateVars

	switch ctx.rinfo.efifile {
	case "shim.efi":
		dir = imageDir
		stageFile = image.ShimFind(imageName)
		if stageFile == "" {
			wwlog.Error("couldn't find shim.efi for %s", imageName)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case "grub.efi", "grub-tpm.efi", "grubx64.efi", "grubia32.efi", "grubaa64.efi", "grubarm.efi":
		dir = imageDir
		stageFile = image.GrubFind(imageName)
		if stageFile == "" {
			wwlog.Error("couldn't find grub*.efi for %s", imageName)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case "grub.cfg":
		dir = path.Join(ctx.conf.Paths.Sysconfdir, "warewulf/grub")
		stageFile = path.Join(dir, "grub.cfg.ww")
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

	sendResponse(w, req, dir, stageFile, tmplData, ctx)
}
