package warewulfd

import (
	"bufio"
	"net/http"
	"os"
	"strings"

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
	req.Body.Close()
	return nil
}

func getOverlayFile(
	n node.NodeInfo,
	context string,
	stage_overlays []string,
	autobuild bool) (stage_file string, err error) {

	stage_file = overlay.OverlayImage(n.Id.Get(), context, stage_overlays)
	err = nil
	build := !util.IsFile(stage_file)
	wwlog.Verbose("stage file: %s", stage_file)
	if !build && autobuild {
		build = util.PathIsNewer(stage_file, nodepkg.ConfigFile)

		for _, overlayname := range stage_overlays {
			build = build || util.PathIsNewer(stage_file, overlay.OverlaySourceDir(overlayname))
		}
	}

	if build {
		err = overlay.BuildOverlay(n, context, stage_overlays)
		if err != nil {
			wwlog.Error("Failed to build overlay: %s, %s, %s\n%s",
				n.Id.Get(), stage_overlays, stage_file, err)
		}
	}

	return
}

var arpFile string

func init() {
	arpFile = "/proc/net/arp"
}

func SetArpFile(newName string) {
	arpFile = newName
}

/*
returns the mac address if it has an entry in the arp cache
*/
func ArpFind(ip string) (mac string) {
	arpCache, err := os.Open(arpFile)
	if err != nil {
		return
	}
	defer arpCache.Close()

	scanner := bufio.NewScanner(arpCache)
	scanner.Scan()
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if strings.EqualFold(fields[0], ip) {
			return fields[3]
		}
	}
	return
}
