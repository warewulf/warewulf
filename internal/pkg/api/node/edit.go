package apinode

import (
	"fmt"

	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"gopkg.in/yaml.v3"
)

/*
Add nodes from yaml
*/
func NodeAddFromYaml(nodeList *wwapiv1.NodeYaml) (err error) {
	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open NodeDB: %w", err)
	}
	nodeMap := make(map[string]*node.Node)
	err = yaml.Unmarshal([]byte(nodeList.NodeConfMapYaml), nodeMap)
	if err != nil {
		return fmt.Errorf("could not unmarshal Yaml: %w", err)
	}
	for nodeName, nodeData := range nodeMap {
		if _, err = nodeDB.GetNodeOnly(nodeName); err == node.ErrNotFound {
			_, err = nodeDB.AddNode(nodeName)
			if err != nil {
				return fmt.Errorf("couldn't add new node: %w", err)
			}
		}
		err = nodeDB.SetNode(nodeName, *nodeData)
		if err != nil {
			return fmt.Errorf("couldn't set node: %w", err)
		}
	}
	err = nodeDB.Persist()
	if err != nil {
		return fmt.Errorf("failed to persist nodedb: %w", err)
	}
	return nil
}
