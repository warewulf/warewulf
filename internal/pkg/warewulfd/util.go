package warewulfd

import (
	"bufio"
	"net/http"
	"os"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
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
	defer func() { _ = fd.Close() }()

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

	wwlog.Info("send %s -> %s", filename, sendto)
	_ = req.Body.Close()
	return nil
}

func getOverlayFile(n node.Node, context string, autobuild bool) (stage_file string, err error) {
	stage_file = overlay.Image(n.Id(), context, nil)
	build := !util.IsFile(stage_file)
	wwlog.Verbose("stage file: %s", stage_file)
	if !build && autobuild {
		build = util.PathIsNewer(stage_file, config.Get().Paths.NodesConf())
	}

	if build {
		registry, err := node.New()
		if err != nil {
			wwlog.Error("Failed to build overlay: %s, %s\n%s",
				n.Id(), stage_file, err)
			return "", err
		}
		var allNodes []node.Node
		allNodes, err = registry.FindAllNodes()
		if err != nil {
			wwlog.Error("Failed to build overlay: %s, %s\n%s",
				n.Id(), stage_file, err)
			return "", err
		}
		err = overlay.BuildAllOverlays([]node.Node{n}, allNodes, 1)
		if err != nil {
			wwlog.Error("Failed to build overlay: %s, %s\n%s",
				n.Id(), stage_file, err)
		}
	}

	return
}

var arpFile string

func init() {
	arpFile = "/proc/net/arp"
}

/*
returns the mac address if it has an entry in the arp cache
*/
func ArpFind(ip string) (mac string) {
	arpCache, err := os.Open(arpFile)
	if err != nil {
		return
	}
	defer func() { _ = arpCache.Close() }()

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
