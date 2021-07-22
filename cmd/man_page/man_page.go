// usage: ./bash_completion <FILE>
package main

import (
	"fmt"	
	"os"
	"github.com/hpcng/warewulf/internal/app/wwctl"
)

func main() {
	
	if err := wwctl.GenManTree(os.Args[1]); err != nil {
		fmt.Println(err)
		return
	}
}
