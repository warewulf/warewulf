// Generate man pages for wwctl command.
// usage: ./man_page <DIRECTORY>
package main

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/app/wwctl"
	"os"
)

func main() {

	if err := wwctl.GenManTree(os.Args[1]); err != nil {
		fmt.Println(err)
		return
	}
}
