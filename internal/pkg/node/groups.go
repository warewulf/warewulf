package node

import (
	"slices"
	"sort"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// Reserved group name that expands to every node not explicitly masked
const AllGroup = "all"

// GroupMembers returns a sorted, deduped list of nodes belonging to the
// named group. Membership is the union of every node and profile that
// declares the group in its `groups:` field. A node carrying the literal
// token "~<name>" (directly or via a profile) is excluded.
func (config *NodesYaml) GroupMembers(name string) []string {
	if name == AllGroup {
		all := config.ListAllNodes()
		filtered := make([]string, 0, len(all))
		for _, id := range all {
			if config.hasLiteralGroup(id, "~"+AllGroup) {
				continue
			}
			filtered = append(filtered, id)
		}
		return filtered
	}

	members := make(map[string]struct{})
	for id := range config.Nodes {
		merged, err := config.GetNode(id)
		if err != nil {
			continue
		}
		if slices.Contains(merged.Groups, name) {
			members[id] = struct{}{}
		}
	}

	// don't warn on declared but empty groups
	if len(members) == 0 {
		declared := false
		for _, g := range config.ListAllGroups() {
			if g == name {
				declared = true
				break
			}
		}
		// print warning if groups is not declared anywhere (aka doesn't exist)
		if !declared {
			wwlog.Warn("unknown group: %s", name)
		}
	}

	result := make([]string, 0, len(members))
	for id := range members {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

// hasLiteralGroup reports whether the node carries the given token verbatim
// in its own groups list or in a non-negated profile's groups list. Used to
// honor `~all` opt-outs, which the normal merge step would otherwise strip.
func (config *NodesYaml) hasLiteralGroup(nodeID, token string) bool {
	if n, ok := config.Nodes[nodeID]; ok {
		if slices.Contains(n.Groups, token) {
			return true
		}
	}
	for _, profileID := range config.getNodeProfiles(nodeID) {
		if strings.HasPrefix(profileID, "~") {
			continue
		}
		if profile, err := config.GetProfile(profileID); err == nil {
			if slices.Contains(profile.Groups, token) {
				return true
			}
		}
	}
	return false
}

// ListNodesUsingGroup returns the merged node objects for every member of
// the named group.
func (config *NodesYaml) ListNodesUsingGroup(name string) ([]Node, error) {
	members := config.GroupMembers(name)
	if len(members) == 0 {
		return nil, nil
	}
	return config.FindAllNodes(members...)
}

// ListAllGroups returns a sorted, deduped list of every group referenced on
// any node or profile. Negated entries (`~name`) are excluded.
func (config *NodesYaml) ListAllGroups() []string {
	seen := make(map[string]struct{})
	for _, node := range config.Nodes {
		for _, g := range node.Groups {
			if strings.HasPrefix(g, "~") {
				continue
			}
			seen[g] = struct{}{}
		}
	}
	for _, profile := range config.NodeProfiles {
		for _, g := range profile.Groups {
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

// Compile-time guard: NodesYaml must satisfy hostlist.GroupResolver.
var _ hostlist.GroupResolver = (*NodesYaml)(nil)
