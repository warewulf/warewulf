package apinode

import (
	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

// nodeDbSave persists the nodeDB to disk and restarts warewulfd.
// TODO: We will likely need locking around anything changing nodeDB
// or restarting warewulfd. Determine if the reason for restart is
// just to reinitialize warewulfd with the new nodeDB or if there is
// something more to it.
func DbSave(nodeDB *node.NodeYaml) (err error) {
	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist nodedb")
	}

	err = warewulfd.DaemonReload()
	if err != nil {
		return errors.Wrap(err, "failed to reload warewulf daemon")
	}
	return
}
