package apiprofile

import (
	"fmt"
	"os"

	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

// NodeSet is the wwapiv1 implmentation for updating nodeinfo fields.
func ProfileSet(set *wwapiv1.NodeSetParameter) (err error) {

	if set == nil {
		return fmt.Errorf("NodeAddParameter is nil")
	}

	var nodeDB node.NodeYaml
	nodeDB, _, err = ProfileSetParameterCheck(set, false)
	if err != nil {
		return errors.Wrap(err, "Could not open database")
	}
	return apinode.DbSave(&nodeDB)
}

// ProfileSetParameterCheck does error checking on ProfileSetParameter.
// Output to the console if console is true.
// TODO: Determine if the console switch does wwlog or not.
// - console may end up being textOutput?
func ProfileSetParameterCheck(set *wwapiv1.NodeSetParameter, console bool) (nodeDB node.NodeYaml, profileCount uint, err error) {

	if set == nil {
		err = fmt.Errorf("Profile set parameter is nil")
		if console {
			fmt.Printf("%v\n", err)
			return
		}
	}

	if set.NodeNames == nil {
		err = fmt.Errorf("Profile set parameter: ProfileNames is nil")
		if console {
			fmt.Printf("%v\n", err)
			return
		}
	}

	nodeDB, err = node.New()
	if err != nil {
		wwlog.Error("Could not open configuration: %s\n", err)
		return
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Error("Could not get profile list: %s\n", err)
		return
	}

	// Note: This does not do expansion on the nodes.

	if set.AllNodes || (len(set.NodeNames) == 0) {
		if console {
			fmt.Printf("\n*** WARNING: This command will modify all profiles! ***\n\n")
		}
	}

	if len(profiles) == 0 {
		if console {
			fmt.Printf("No profiles found\n")
		}
		return
	}

	for _, p := range profiles {
		if util.InSlice(set.NodeNames, p.Id.Get()) {
			wwlog.Printf(wwlog.VERBOSE, "Evaluating profile: %s\n", p.Id.Get())
			for key, val := range set.OptionsStrMap {
				if val != "" {
					wwlog.Verbose("profile:%s setting %s to %s\n", p.Id.Get(), key, val)
					p.SetField(key, val)
				}
			}

			if set.NetdevDelete != "" {

				if _, ok := p.NetDevs[set.NetdevDelete]; !ok {
					err = fmt.Errorf("Network device name doesn't exist: %s", set.NetdevDelete)
					wwlog.Error(fmt.Sprintf("%v\n", err.Error()))
					return
				}

				wwlog.Printf(wwlog.VERBOSE, "Profile: %s, Deleting network device: %s\n", p.Id.Get(), set.NetdevDelete)
				delete(p.NetDevs, set.NetdevDelete)
			}

			err := nodeDB.ProfileUpdate(p)
			if err != nil {
				wwlog.Error("%s\n", err)
				os.Exit(1)
			}

			profileCount++
		}
	}
	return
}

/*
Adds a new profile with the given name
*/
func AddProfile(set *wwapiv1.NodeSetParameter, console bool) error {
	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "Could not open database")
	}

	if util.InSlice(nodeDB.ListAllProfiles(), set.NodeNames[0]) {
		return errors.New(fmt.Sprintf("profile with name %s allready exists", set.NodeNames[0]))
	}
	_, err = nodeDB.AddProfile(set.NodeNames[0])
	if err != nil {
		return errors.Wrap(err, "Could not create new profile")
	}
	err = apinode.DbSave(&nodeDB)
	if err != nil {
		return errors.Wrap(err, "Could not persist new profile")
	}
	return nil
}