package apiprofile

import (
	"fmt"
	"os"

	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

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
		return fmt.Errorf("couldn't open NodeDB: %w", err)
	}
	if nodeDB.StringHash() != nodeList.Hash && !nodeList.Force {
		return fmt.Errorf("got wrong hash, not modifying profile database")
	}

	profileMap := make(map[string]*node.ProfileConf)
	err = yaml.Unmarshal([]byte(nodeList.NodeConfYaml), profileMap)
	if err != nil {
		return fmt.Errorf("couldn't unmarshall Yaml: %w", err)
	}
	for profileName, profile := range profileMap {
		err = nodeDB.SetProfile(profileName, *profile)
		if err != nil {
			return fmt.Errorf("couldn't set profile: %w", err)
		}
	}
	err = nodeDB.Persist()
	if err != nil {
		return fmt.Errorf("failed to persist nodedb: %w", err)
	}
	return nil
}
