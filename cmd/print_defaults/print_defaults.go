package main

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/node"
)

/*
Print the build in defaults for the nodes.
Called via Makefile so that there is single upstream
source of the defaults which is FallBackConf
*/

func main() {
	fmt.Println(node.FallBackConf)
}
