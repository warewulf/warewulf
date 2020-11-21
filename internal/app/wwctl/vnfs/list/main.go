package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	config := config.New()

	files, err := ioutil.ReadDir(config.VnfsImageParentDir())
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}

	fmt.Printf("VNFS LIST: work in progress: %s\n", config.VnfsImageParentDir())
	return nil
}