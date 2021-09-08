package main

import (
	"os"

	"github.com/hpcng/warewulf/internal/app/wwctl"
)

func main() {
	root := wwctl.GetRootCommand()

	err := root.Execute()
	if err != nil {
		os.Exit(255)
	}
}
