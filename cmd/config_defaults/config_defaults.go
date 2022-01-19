package main

// Regenerates the default configuration files
// Keeps the current content and adds missing default values
// Won't create a new file, but will update a blank file

//TODO: Add nodes.conf

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

func main() {
	tmpConf, err := warewulfconf.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tmpConf.Persist()
	if err != nil {
		fmt.Println(err)
		return
	}
}
