package main

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/app/wwclient"
)

func main() {

	root := wwclient.GetRootCommand()

	err := root.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		if wwclient.DebugFlag {
			fmt.Printf("\nSTACK TRACE: %+v\n", err)
		}
		os.Exit(255)
	}
}
