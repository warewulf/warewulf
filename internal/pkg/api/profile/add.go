package apiprofile

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"gopkg.in/yaml.v3"
)

/*
Adds a new profile with the given name
*/
func ProfileAdd(nsp *wwapiv1.NodeAddParameter) error {
	if nsp == nil {
		return fmt.Errorf("NodeAddParameter is nill")
	}
	nodeDB, err := node.New()
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	for _, p := range nsp.NodeNames {
		if util.InSlice(nodeDB.ListAllProfiles(), p) {
			return errors.New(fmt.Sprintf("profile with name %s already exists", p))
		}
		pNew, err := nodeDB.AddProfile(p)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal([]byte(nsp.NodeConfYaml), &pNew)
		if err != nil {
			return fmt.Errorf("failed to add profile: %w", err)
		}
	}
	err = nodeDB.Persist()
	if err != nil {
		return fmt.Errorf("failed to persist new profile: %w", err)
	}
	return nil
}
