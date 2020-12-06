package pull

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var name string
	uri := args[0]

	if len(args) == 2 {
		name = args[1]
	} else {
		name = path.Base(uri)
		fmt.Printf("Setting VNFS name: %s\n", name)
	}

	if vnfs.ValidName(name) == false {
		wwlog.Printf(wwlog.ERROR, "VNFS name contains illegal characters: %s\n", name)
		os.Exit(1)
	}

	fullPath := vnfs.SourceDir(name)

	if util.IsDir(fullPath) == true {
		if SetForce == true {
			wwlog.Printf(wwlog.WARN, "Overwriting existing VNFS\n")
			err := os.RemoveAll(fullPath)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		} else if SetUpdate == true {
			wwlog.Printf(wwlog.WARN, "Updating existing VNFS\n")
		} else {
			wwlog.Printf(wwlog.ERROR, "VNFS Name exists, specify --force, --update, or choose a different name: %s\n", name)
			os.Exit(1)
		}
	}

	err := vnfs.PullURI(uri, name)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not pull image: %s\n", err)
		os.Exit(1)
	}

	return nil
}
