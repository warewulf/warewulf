package main

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/app/wwctl"
)

func main() {
	root := wwctl.GetRootCommand()

	err := root.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		if wwctl.DebugFlag {
			fmt.Printf("\nSTACK TRACE: %+v\n", err)
		}
		os.Exit(255)
	}
}
