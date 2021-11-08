package warewulfd

import (
	"encoding/json"
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/node"
)

type nodeDB struct {
	lock     sync.RWMutex
	NodeInfo map[string]node.NodeInfo
}

var (
	db nodeDB
)

func LoadNodeDB() error {
	TmpMap := make(map[string]node.NodeInfo)

	daemonLogf("(re)Loading the node Database\n")

	DB, err := node.New()
	if err != nil {
		return err
	}

	nodes, err := DB.FindAllNodes()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		for _, netdev := range n.NetDevs {
			TmpMap[netdev.Hwaddr.Get()] = n
		}
	}

	db.lock.Lock()
	defer db.lock.Unlock()
	db.NodeInfo = TmpMap

	return nil
}

func GetNode(val string) (node.NodeInfo, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if _, ok := db.NodeInfo[val]; ok {

		return db.NodeInfo[val], nil
	}

	var empty node.NodeInfo
	return empty, errors.New("No node found")
}

// NodeList for status information
type NodeSlice []*node.NodeInfo

// Len is part of sort.Interface.
func (d NodeSlice) Len() int { return len(d) }

// Swap is part of sort.Interface.
func (d NodeSlice) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

// Less is part of sort.Interface. We use count as the value to sort by
func (d NodeSlice) Less(i, j int) bool { return d[i].LastSeen < d[j].LastSeen }

type NodeList map[string]*node.NodeInfo

func (n NodeList) AddTestNodes(c int) {
	now := time.Now().Unix()
	rand.Seed(now)
	for i := 0; i < c; i++ {
		name := fmt.Sprintf("test_%d", i)
		offset := now - int64(rand.Intn(360))
		n[name] = &node.NodeInfo{}
		n[name].Id.Set(name)
		n[name].ClusterName.Set("test")
		n[name].LastSeen = offset
	}
}

func (n NodeList) Sort() []*node.NodeInfo {
	s := make(NodeSlice, 0, len(n))

	for _, d := range n {
		s = append(s, d)
	}
	sort.Sort(s)

	ret := []*node.NodeInfo{}

	for _, v := range s {
		ret = append(ret, v)
	}

	return ret
}

func (n NodeList) JsonSend(w http.ResponseWriter) {
	jsonString, err := json.Marshal(NodeInfoDB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonString)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "failed to write json data %s", err)
	}
}
