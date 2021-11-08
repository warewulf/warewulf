package warewulfd

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

// TODO: move to separate file?
const nodeStatusHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
<style>
table {
  font-family: arial, sans-serif;
  border-collapse: collapse;
  width: 100%;
}

td, th {
  border: 1px solid #dddddd;
  text-align: left;
  padding: 8px;
}

tr:nth-child(even) {
  background-color: #dddddd;
}
</style>
</head>
<body>
    <h1>{{.PageTitle}}</h1>
    <table>
      <tr>
        <th>Node</th>
        <th>Cluster</th>
        <th>last seen (s)</th>
      </tr>
        {{range .HtmlBody}}
            <tr>
                <td>{{.Node}}</td>
                <td>{{.Cluster}}</td>
                <td>{{.LastSeen}}</td>
            </tr>
        {{end}}
    </table>
</body>
</html>`

var NodeInfoDB = make(NodeList)

type TemplData struct {
	Node     string
	Cluster  string
	LastSeen int64
}

type PageData struct {
	PageTitle string
	HtmlBody  []TemplData
}

func NodeStatusSend(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Server", "Warewulf status information")
	// query handling
	q := req.URL.Query()
	if q.Has("test") {
		count, err := strconv.Atoi(q["test"][0])
		if err == nil {
			NodeInfoDB.AddTestNodes(count)
		}
	}
	if q.Has("json") {
		NodeInfoDB.JsonSend(w)
		return
	}
	// human readable
	maxLines := len(NodeInfoDB)
	if q.Has("limit") {
		l, err := strconv.Atoi(q["limit"][0])
		if err == nil {
			maxLines = l
		}
	}

	// Make and parse the HTML template
	t, err := template.New("webpage").Parse(nodeStatusHtmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := PageData{
		PageTitle: "node status info",
		HtmlBody:  []TemplData{},
	}

	// body
	now := time.Now().UTC().Unix()
	sorted := NodeInfoDB.Sort()
	i := 0
	for _, n := range sorted {
		i++
		if i > maxLines {
			break
		}
		data.HtmlBody = append(data.HtmlBody, TemplData{n.Id.Get(), n.ClusterName.Get(), now - n.LastSeen})
	}
	// render and write
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func RuntimeOverlaySend(w http.ResponseWriter, req *http.Request) {
	conf, err := warewulfconf.New()
	if err != nil {
		daemonLogf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	nodes, err := node.New()
	if err != nil {
		daemonLogf("ERROR: Could not read node configuration file: %s\n", err)
		w.WriteHeader(503)
		return
	}

	remote := strings.Split(req.RemoteAddr, ":")
	port, err := strconv.Atoi(remote[1])
	if err != nil {
		daemonLogf("ERROR: Could not convert port to integer: %s\n", remote[1])
		w.WriteHeader(503)
		return
	}

	if err != nil {
		daemonLogf("ERROR: Could not load configuration file: %s\n", err)
		return
	}

	if conf.Warewulf.Secure {
		if port >= 1024 {
			daemonLogf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
			w.WriteHeader(401)
			return
		}
	}

	n, err := nodes.FindByIpaddr(remote[0])
	if err != nil {
		daemonLogf("WARNING: Could not find node by IP address: %s\n", remote[0])
		w.WriteHeader(404)
		return
	}

	if !n.Id.Defined() {
		daemonLogf("REQ:   %15s: %s (unknown/unconfigured node)\n", n.Id.Get(), req.URL.Path)
		w.WriteHeader(404)
		return
	} else {
		daemonLogf("REQ:   %15s: %s\n", n.Id.Get(), req.URL.Path)
		n.LastSeen = time.Now().UTC().Unix()
		NodeInfoDB[n.Id.Get()] = &n
	}

	if n.RuntimeOverlay.Defined() {
		fileName := config.RuntimeOverlayImage(n.Id.Get())

		if conf.Warewulf.AutobuildOverlays {
			if !util.IsFile(fileName) || util.PathIsNewer(fileName, node.ConfigFile) || util.PathIsNewer(fileName, config.RuntimeOverlaySource(n.RuntimeOverlay.Get())) {
				daemonLogf("BUILD: %15s: Runtime Overlay\n", n.Id.Get())
				_ = overlay.BuildRuntimeOverlay([]node.NodeInfo{n})
			}
		}

		err := sendFile(w, fileName, n.Id.Get())
		if err != nil {
			daemonLogf("ERROR: %s\n", err)
		}
	} else {
		w.WriteHeader(503)
		daemonLogf("WARNING: No 'runtime system-overlay' set for node %s\n", n.Id.Get())
	}
}
