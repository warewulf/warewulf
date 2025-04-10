package node

import "slices"

// ListAllNodes returns a slice of all node names defined in the Nodes map.
func (config *NodesYaml) ListAllNodes() []string {
	nodes := []string{}
	for n := range config.Nodes {
		nodes = append(nodes, n)
	}
	slices.Sort(nodes)
	return nodes
}

// ListAllProfiles returns a slice of all profile IDs defined in the
// NodeProfiles map.
func (config *NodesYaml) ListAllProfiles() []string {
	profiles := []string{}
	for p := range config.NodeProfiles {
		profiles = append(profiles, p)
	}
	slices.Sort(profiles)
	return profiles
}

// ListNodesUsingProfile returns a slice of node IDs that reference the
// specified profileID.
func (config *NodesYaml) ListNodesUsingProfile(profileID string) []string {
	nodes := []string{}
	for n := range config.Nodes {
		if slices.Contains(config.Nodes[n].Profiles, profileID) {
			nodes = append(nodes, n)
		}
	}
	slices.Sort(nodes)
	return nodes
}

// ListProfilesUsingProfile returns a slice of profile IDs from NodeProfiles
// that reference the specified profileID.
func (config *NodesYaml) ListProfilesUsingProfile(profileID string) []string {
	profiles := []string{}
	for p := range config.NodeProfiles {
		if slices.Contains(config.NodeProfiles[p].Profiles, profileID) {
			profiles = append(profiles, p)
		}
	}
	slices.Sort(profiles)
	return profiles
}

// ListNodesUsingImage returns a slice of node IDs for nodes that use the
// specified image.
func (config *NodesYaml) ListNodesUsingImage(image string) []string {
	nodes := []string{}
	for n := range config.Nodes {
		if config.Nodes[n].ImageName == image {
			nodes = append(nodes, n)
		}
	}
	slices.Sort(nodes)
	return nodes
}

// ListProfilesUsingImage returns a slice of profile IDs for profiles that use
// the specified image.
func (config *NodesYaml) ListProfilesUsingImage(image string) []string {
	profiles := []string{}
	for p := range config.NodeProfiles {
		if config.NodeProfiles[p].ImageName == image {
			profiles = append(profiles, p)
		}
	}
	slices.Sort(profiles)
	return profiles
}

// ListNodesUsingOverlay returns a slice of node IDs for nodes that include the
// specified overlay in either RuntimeOverlay or SystemOverlay.
func (config *NodesYaml) ListNodesUsingOverlay(overlay string) []string {
	nodes := []string{}
	for n := range config.Nodes {
		if slices.Contains(config.Nodes[n].RuntimeOverlay, overlay) ||
			slices.Contains(config.Nodes[n].SystemOverlay, overlay) {
			nodes = append(nodes, n)
		}
	}
	slices.Sort(nodes)
	return nodes
}

// ListProfilesUsingOverlay returns a slice of profile IDs for profiles that
// include the specified overlay in either RuntimeOverlay or SystemOverlay.
func (config *NodesYaml) ListProfilesUsingOverlay(overlay string) []string {
	profiles := []string{}
	for p := range config.NodeProfiles {
		if slices.Contains(config.NodeProfiles[p].RuntimeOverlay, overlay) ||
			slices.Contains(config.NodeProfiles[p].SystemOverlay, overlay) {
			profiles = append(profiles, p)
		}
	}
	slices.Sort(profiles)
	return profiles
}
