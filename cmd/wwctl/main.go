package main

import (
	"github.com/hpcng/warewulf/internal/app/wwctl"
)

func main() {
	root := wwctl.GetRootCommand()

	root.Execute()
}

