package main

import (
	"github.com/hpcng/warewulf/internal/app/warewulfd"
)


func main() {
	root := warewulfd.GetRootCommand()

	root.Execute()
}