package imprt

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	apinode "github.com/warewulf/warewulf/internal/pkg/api/node"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"gopkg.in/yaml.v3"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	file, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("could not open file: %s", err)
	}
	defer file.Close()

	importMap := make(map[string]*node.Node)
	buffer, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("could not read: %s", err)
	}

	err = yaml.Unmarshal(buffer, importMap)
	if err != nil {
		return fmt.Errorf("could not parse import file: %s", err)
	}

	if setYes || util.Confirm(fmt.Sprintf("Are you sure you want to modify %d nodes", len(importMap))) {
		err = apinode.NodeAddFromYaml(&wwapiv1.NodeYaml{NodeConfMapYaml: string(buffer)})
		if err != nil {
			return fmt.Errorf("got following problem when writing back yaml: %s", err)
		}
	}

	return nil
}
