package node

import (
	"slices"
	"sort"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// Rserved group name that expands to every node not explicitly masked
const AllGroup = "all"

// Returns sorted, deduped list. Union of nodesgroups from nodes.conf,
// per-node or profile
func (config *NodesYaml) GroupMembers(name string) []string {
	if name == AllGroup {
		all := config.ListAllNodes()
		filtered := make([]string, 0, len(all))
		// mask node if it has an "~all", explicit or inherited 
		for _, id := range all {
			if config.hasLiteralNodegroup(id, "~"+AllGroup) {
				continue
			}
			filtered = append(filtered, id)
		}
		return filtered
	}

	members := make(map[string]struct{})

	for _, entry := range hostlist.Expand(config.NodeGroups[name]) {
		if _, ok := config.Nodes[entry]; ok {
			members[entry] = struct{}{}
		} else {
			wwlog.Warn("nodegroup %q references unknown node: %s", name, entry)
		}
	}

	for id := range config.Nodes {
		merged, err := config.GetNode(id)
		if err != nil {
			continue
		}
		if slices.Contains(merged.NodeGroups, name) {
			members[id] = struct{}{}
		}
	}

	if len(members) == 0 {
		// Only warn if neither source mentioned the nodegroup; otherwise the
		// caller asked for a defined-but-empty nodegroup, which is fine.
		if _, defined := config.NodeGroups[name]; !defined {
			wwlog.Warn("unknown nodegroup: %s", name)
		}
	}

	result := make([]string, 0, len(members))
	for id := range members {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

// Check for token regardless of ~ negation
func (config *NodesYaml) hasLiteralNodegroup(nodeID, token string) bool {
	if n, ok := config.Nodes[nodeID]; ok {
		if slices.Contains(n.NodeGroups, token) {
			return true
		}
	}
	for _, profileID := range config.getNodeProfiles(nodeID) {
		if strings.HasPrefix(profileID, "~") {
			continue
		}
		if profile, err := config.GetProfile(profileID); err == nil {
			if slices.Contains(profile.NodeGroups, token) {
				return true
			}
		}
	}
	return false
}

// Returns node objects for every nodegroup member
func (config *NodesYaml) ListNodesUsingNodegroup(name string) ([]Node, error) {
	members := config.GroupMembers(name)
	if len(members) == 0 {
		return nil, nil
	}
	return config.FindAllNodes(members...)
}

// Returns a sorted/deduped list of every nodegroup from nodes.conf, node or profile.
func (config *NodesYaml) ListAllNodegroups() []string {
	seen := make(map[string]struct{})
	for name := range config.NodeGroups {
		seen[name] = struct{}{}
	}
	for _, node := range config.Nodes {
		for _, g := range node.NodeGroups {
			if strings.HasPrefix(g, "~") {
				continue
			}
			seen[g] = struct{}{}
		}
	}
	for _, profile := range config.NodeProfiles {
		for _, g := range profile.NodeGroups {
			if strings.HasPrefix(g, "~") {
				continue
			}
			seen[g] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}
