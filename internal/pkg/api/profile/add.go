package apiprofile

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
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
		return errors.Wrap(err, "Could not open database")
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
			return errors.Wrap(err, "failed to add profile")
		}
	}
	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist new profile")
	}
	return nil
}
