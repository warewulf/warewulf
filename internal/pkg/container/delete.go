package container

import (
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type DeleteParameter struct {
	Names []string
}

func Delete(param *DeleteParameter) error {
	db, err := node.New()
	if err != nil {
		return err
	}

	nodes, err := db.FindAllNodes()
	if err != nil {
		return err
	}

ARG_LOOP:
	for _, name := range param.Names {
		for _, n := range nodes {
			if n.ContainerName.Get() == name {
				wwlog.Error("Container is configured for nodes, skipping: %s", name)
				continue ARG_LOOP
			}
		}

		if !ValidSource(name) {
			wwlog.Error("Container name is not a valid source: %s", name)
			continue
		}
		err := DeleteSource(name)
		if err != nil {
			wwlog.Error("Could not remove source: %s", name)
		}
		err = DeleteImage(name)
		if err != nil {
			wwlog.Error("Could not remove image files %s", name)
		}
	}

	return nil
}
