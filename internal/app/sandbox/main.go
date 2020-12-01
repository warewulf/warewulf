package main

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
)

func main() {
	nodeDB, _ := node.New()
	nodes, _ := nodeDB.FindAllNodes()
//	wwlog.SetLevel(wwlog.DEBUG)
	overlay.OverlayBuild(nodes, "runtime")
}
