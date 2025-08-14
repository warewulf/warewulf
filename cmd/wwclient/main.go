package main

import (
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/app/wwclient"
)

func main() {

	root := wwclient.GetRootCommand()

	err := root.Execute()
	if err != nil {
		if wwclient.DebugFlag {
			fmt.Printf("\nSTACK TRACE: %+v\n", err)
		}
		os.Exit(255)
	}
}
