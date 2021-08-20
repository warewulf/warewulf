// usage: ./bash_completion <FILE>
package main

import (
	"fmt"
	"os"
	"github.com/hpcng/warewulf/internal/app/wwctl"
)

func main() {
	fh, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	defer fh.Close()

	if err := wwctl.GenBashCompletion(fh); err != nil {
		fmt.Println(err)
		return
	}
}
