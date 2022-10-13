package apiprofile

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
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
	profileMap := nodeDB.NodeProfiles
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
	profileMap := nodeDB.NodeProfiles
	profileMap = node.FilterMapByName(profileMap, profileList.Output)
	buffer, _ := yaml.Marshal(profileMap)
	retVal := wwapiv1.NodeYaml{
		NodeConfMapYaml: string(buffer),
	}
	return &retVal
}

/*
Add profiles from yaml
*/
func ProfileAddFromYaml(nodeList *wwapiv1.NodeYaml) (err error) {
	nodeDB, err := node.New()
	if err != nil {
		return errors.Wrap(err, "Could not open NodeDB: %s\n")
	}
	profileMap := make(map[string]*node.NodeConf)
	err = yaml.Unmarshal([]byte(nodeList.NodeConfMapYaml), profileMap)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshall Yaml: %s\n")
	}
	for profileName, profile := range profileMap {
		nodeDB.NodeProfiles[profileName] = profile
	}
	err = nodeDB.Persist()
	if err != nil {
		return errors.Wrap(err, "failed to persist nodedb")
	}
	return nil
}
