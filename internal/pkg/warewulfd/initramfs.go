package warewulfd

import (
	"net/http"

	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// HandleInitramfs handles initramfs requests
func HandleInitramfs(w http.ResponseWriter, req *http.Request) {
	ctx, err := initHandleRequest(w, req)
	if err != nil {
		return // response already written
	}

	var stageFile string

	if !ctx.remoteNode.Valid() {
		wwlog.Error("%s (unknown/unconfigured node)", ctx.rinfo.hwaddr)
	} else {
		if kernel_ := kernel.FromNode(&ctx.remoteNode); kernel_ != nil {
			if kver := kernel_.Version(); kver != "" {
				if initramfs := image.FindInitramfs(ctx.remoteNode.ImageName, kver); initramfs != nil {
					stageFile = initramfs.FullPath()
				} else {
					wwlog.Error("No initramfs found for kernel %s in image %s", kver, ctx.remoteNode.ImageName)
				}
			} else {
				wwlog.Error("No initramfs found: unable to determine kernel version for node %s", ctx.remoteNode.Id())
			}
		} else {
			wwlog.Error("No initramfs found: unable to find kernel for node %s", ctx.remoteNode.Id())
		}
	}

	sendResponse(w, req, stageFile, nil, ctx)
}
