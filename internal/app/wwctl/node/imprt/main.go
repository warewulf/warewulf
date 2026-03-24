package imprt

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"gopkg.in/yaml.v3"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	file, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("could not open file: %s", err)
	}
	defer func() { _ = file.Close() }()

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
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open NodeDB: %w", err)
		}
		for nodeName, nodeData := range importMap {
			if _, err := nodeDB.GetNodeOnly(nodeName); err == node.ErrNotFound {
				if _, err := nodeDB.AddNode(nodeName); err != nil {
					return fmt.Errorf("couldn't add new node: %w", err)
				}
			}
			if err := nodeDB.SetNode(nodeName, *nodeData); err != nil {
				return fmt.Errorf("couldn't set node: %w", err)
			}
		}
		if err := nodeDB.Persist(); err != nil {
			return fmt.Errorf("failed to persist nodedb: %w", err)
		}
	}

	return nil
}
