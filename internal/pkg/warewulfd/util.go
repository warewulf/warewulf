package warewulfd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	nodepkg "github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func sendFile(
	w http.ResponseWriter,
	req *http.Request,
	filename string,
	sendto string) error {

	fd, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	http.ServeContent(
		w,
		req,
		filename,
		stat.ModTime(),
		fd)

	wwlog.Send("%15s: %s", sendto, filename)

	return nil
}

func getOverlayFile(
	nodeId string,
	context string,
	stage_overlays []string,
	autobuild bool) (stage_file string, err error) {

	stage_file = overlay.OverlayImage(nodeId, context, stage_overlays)
	err = nil

	build := !util.IsFile(stage_file)

	if !build && autobuild {
		build = util.PathIsNewer(stage_file, nodepkg.ConfigFile)

		for _, overlayname := range stage_overlays {
			build = build || util.PathIsNewer(stage_file, overlay.OverlaySourceDir(overlayname))
		}
	}

	if build {
		nodeDB, errNested := node.New()
		if err != nil {
			return stage_file, errNested
		}
		myNode, errNested := nodeDB.FindAllNodes()
		if err != nil {
			return stage_file, errNested
		}
		myNode = node.FilterByName(myNode, []string{nodeId})
		if len(myNode) != 1 {
			return stage_file, fmt.Errorf("couldn't find node %s", nodeId)
		}
		err = overlay.BuildOverlay(myNode[0], context, stage_overlays)
		if err != nil {
			wwlog.Error("Failed to build overlay: %s, %s, %s\n%s",
				nodeId, stage_overlays, stage_file, err)
		}
	}

	return
}
