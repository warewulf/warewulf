// usage: ./bash_completion <FILE>
package main

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/app/wwctl"
	"os"
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
