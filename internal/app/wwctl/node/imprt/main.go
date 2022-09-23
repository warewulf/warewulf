package imprt

import (
	"fmt"
	"io"
	"os"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	apiutil "github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	file, err := os.Open(args[0])
	if err != nil {
		wwlog.Error("Could not open file:%s \n", err)
		os.Exit(1)
	}
	defer file.Close()

	importMap := make(map[string]*node.NodeConf)
	buffer, err := io.ReadAll(file)
	if err != nil {
		wwlog.Error("Could not read:%s\n", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(buffer, importMap)
	if err == nil {
		yes := apiutil.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes", len(importMap)))
		if yes {
			err = apinode.NodeAddFromYaml(&wwapiv1.NodeYaml{NodeConfMapYaml: string(buffer)})
			if err != nil {
				wwlog.Error("Got following problem when writing back yaml: %s", err)
				os.Exit(1)
			}
		}
	} else {
		wwlog.Error("Could not parse import file")
	}

	return nil
}
