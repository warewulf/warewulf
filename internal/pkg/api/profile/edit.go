package apiprofile

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

/*
Returns the nodes as a yaml string
*/
func FindAllProfileConfs() *wwapiv1.NodeYaml {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open nodeDB: %s\n", err)
		os.Exit(1)
	}
	profileMap, _ := nodeDB.FindAllProfiles()
	// ignore err as nodeDB should always be correct
	buffer, _ := yaml.Marshal(profileMap)
	retVal := wwapiv1.NodeYaml{
		NodeConfMapYaml: string(buffer),
	}
	return &retVal
}

/*
Returns filtered list of nodes
*/
func FilteredProfiles(profileList *wwapiv1.NodeList) *wwapiv1.NodeYaml {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open nodeDB: %s\n", err)
		os.Exit(1)
	}
	profiles, _ := nodeDB.FindAllProfiles()
	profiles = node.FilterProfileListByName(profiles, profileList.Output)
	buffer, _ := yaml.Marshal(profiles)
	retVal := wwapiv1.NodeYaml{
		NodeConfMapYaml: string(buffer),
	}
	return &retVal
}

/*
Add profiles from yaml
*/
func ProfileAddFromYaml(nodeList *wwapiv1.NodeAddParameter) (err error) {
	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "Could not open NodeDB: %s\n")
	}
	if nodeDB.StringHash() != nodeList.Hash && !nodeList.Force {
		return fmt.Errorf("got wrong hash, not modifying profile database")
	}

	profileMap := make(map[string]*node.ProfileConf)
	err = yaml.Unmarshal([]byte(nodeList.NodeConfYaml), profileMap)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshall Yaml: %s\n")
	}
	for profileName, profile := range profileMap {
		err = nodeDB.SetProfile(profileName, *profile)
		if err != nil {
			return errors.Wrap(err, "couldn't set profile")
		}
	}
	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist nodedb")
	}
	return nil
}
