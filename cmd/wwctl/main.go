package main

import (
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/app/wwctl"
)

func main() {
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
