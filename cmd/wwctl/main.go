package main

import (
	"fmt"
	"os"

	"github.com/containers/storage/pkg/reexec"
	"github.com/warewulf/warewulf/internal/app/wwctl"
)

func main() {
	if reexec.Init() {
		return
	}

	root := wwctl.GetRootCommand()

	err := root.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		if wwctl.DebugFlag {
			fmt.Printf("\nSTACK TRACE: %+v\n", err)
		}
		os.Exit(255)
	}
}
