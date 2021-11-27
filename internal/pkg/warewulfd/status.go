package warewulfd

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"html/template"

	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// NodeList for status information
type NodeStatus struct {
	n *node.NodeInfo
	t int64
}

type NodeSlice []*NodeStatus

// Len is part of sort.Interface.
func (d NodeSlice) Len() int { return len(d) }

// Swap is part of sort.Interface.
func (d NodeSlice) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

// Less is part of sort.Interface. We use count as the value to sort by
func (d NodeSlice) Less(i, j int) bool { return d[i].t < d[j].t }

type NodeList map[string]*NodeStatus

func (n NodeList) AddTestNodes(c int) {
	now := time.Now().Unix()
	rand.Seed(now)
	for i := 0; i < c; i++ {
		name := fmt.Sprintf("test_%d", i)
		offset := now - int64(rand.Intn(360))
		testNode := &node.NodeInfo{}
		testNode.Id.Set(name)
		testNode.ClusterName.Set("test")
		n[name] = &NodeStatus{testNode, offset}
	}
}

func (n NodeList) Sort() []*NodeStatus {
	s := make(NodeSlice, 0, len(n))

	for _, d := range n {
		s = append(s, d)
	}
	sort.Sort(s)

	ret := []*NodeStatus{}

	for _, v := range s {
		ret = append(ret, v)
	}

	return ret
}

// date for html render
type TemplData struct {
	Node     string
	Cluster  string
	LastSeen int64
}

type PageData struct {
	PageTitle string
	HtmlBody  []TemplData
}

var NodeInfoDB = make(NodeList)

func NodeStatusSend(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Server", "Warewulf status information")
	// query handling
	q := req.URL.Query()
	if q.Has("test") {
		count, err := strconv.Atoi(q["test"][0])
		wwlog.Printf(wwlog.INFO, "creating %d test nodes", count)
		if err == nil {
			NodeInfoDB.AddTestNodes(count)
		}
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
	sorted := NodeInfoDB.Sort()
	now := time.Now().UTC().Unix()
	i := 0
	for _, n := range sorted {
		i++
		if i > maxLines {
			break
		}
		data.HtmlBody = append(data.HtmlBody, TemplData{n.n.Id.Get(), n.n.ClusterName.Get(), now - n.t})
	}

	// render and write
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
