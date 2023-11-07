package apiprofile

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfd"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/pkg/errors"
)

// ProfileDelete adds profile deletion for management by Warewulf.
func ProfileDelete(ndp *wwapiv1.NodeDeleteParameter) (err error) {

	var profileList []node.NodeInfo
	profileList, err = ProfileDeleteParameterCheck(ndp, false)
	if err != nil {
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Failed to open node database: %s\n", err)
		return
	}
	if nodeDB.StringHash() != ndp.Hash && !ndp.Force {
		return fmt.Errorf("got wrong hash, not modifying profile database")
	}
	for _, p := range profileList {
		err := nodeDB.DelProfile(p.Id.Get())
		if err != nil {
			wwlog.Error("%s\n", err)
		} else {
			//count++
			wwlog.Verbose("Deleting profile: %s\n", p.Id.Print())
		}
	}

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

// ProfileDeleteParameterCheck does error checking on ProfileDeleteParameter.
// Output to the console if console is true.
// Returns the profiles to delete.
func ProfileDeleteParameterCheck(ndp *wwapiv1.NodeDeleteParameter, console bool) (profileList []node.NodeInfo, err error) {

	if ndp == nil {
		err = fmt.Errorf("ProfileDeleteParameter is nil")
		return
	}

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Failed to open node database: %s\n", err)
		return
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Error("Could not get node list: %s\n", err)
		return
	}

	node_args := hostlist.Expand(ndp.NodeNames)

	for _, r := range node_args {
		var match bool
		for _, p := range profiles {
			if p.Id.Get() == r {
				profileList = append(profileList, p)
				match = true
			}
		}

		if !match {
			wwlog.Error("No match for node: %s\n", r)
		}
	}

	if len(profileList) == 0 {
		wwlog.Info("No s found")
	}
	return
}
