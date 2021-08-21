// usage: ./bash_completion <FILE>
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
