package hostlist

import (
	"fmt"
	"strconv"
	"strings"
)

const Docstring = "Node patterns are a comma-separated list of individual patterns.\nEach pattern can either be a full node name or a node range like node[01-03,05]."

// GroupPrefix marks a token passed to Expand as a node-group reference rather
// than a literal host name or bracketed range, e.g. "@rack1".
const GroupPrefix = "@"

// GroupResolver expands a single group name (without the leading "@")
type GroupResolver interface {
	GroupMembers(name string) []string
}

// Groups without members return as empty
type nopResolver struct{}

func (nopResolver) GroupMembers(string) []string { return nil }

var groupResolver GroupResolver = nopResolver{}

// Install resolver for groups unless nil
func SetGroupResolver(r GroupResolver) {
	if r == nil {
		r = nopResolver{}
	}
	groupResolver = r
}

// Expand takes a slice of host strings, possibly containing comma-separated
// values and bracketed ranges (e.g. "node[01-03]"), and returns a fully
// expanded slice of host names.
//
// Tokens prefixed with "@" are treated as node-group references and resolved
// via the resolver registered with SetGroupResolver; the union of all
// resolved members is returned, deduplicated against any plain host names in
// the same call. Before any resolver is registered the default no-op
// resolver is active, so groups without members are dropped.
func Expand(list []string) []string {
	// First, split each input string on commas that occur outside brackets.
	var preList []string
	for _, s := range list {
		parts := splitTopLevel(s)
		preList = append(preList, parts...)
	}

	expanded := preList
	for {
		onceExpanded, count := expandOnce(expanded)
		if count == 0 {
			break
		}
		expanded = onceExpanded
	}

	return resolveGroups(expanded)
}

// resolveGroups walks bracket-expanded token. Plain tokens pass
// through in order; groups append their members at the position of "@".
// Duplicates are removed.
func resolveGroups(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	var result []string
	add := func(s string) {
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		result = append(result, s)
	}
	for _, tok := range tokens {
		if !strings.HasPrefix(tok, GroupPrefix) {
			add(tok)
			continue
		}
		name := strings.TrimPrefix(tok, GroupPrefix)
		if name == "" {
			continue
		}
		for _, id := range groupResolver.GroupMembers(name) {
			add(id)
		}
	}
	return result
}

// expandOnce performs a single round of bracket expansion.
func expandOnce(hosts []string) ([]string, int) {
	var result []string
	var expansionCount int

	for _, host := range hosts {
		bracketStart := strings.Index(host, "[")
		bracketEnd := strings.Index(host, "]")

		if bracketStart >= 0 && bracketStart < bracketEnd {
			prefix := host[:bracketStart]
			suffix := host[bracketEnd+1:]
			// Extract the content between brackets and split on commas.
			ranges := strings.Split(host[bracketStart+1:bracketEnd], ",")
			expansionCount++

			for _, rng := range ranges {
				parts := strings.Split(rng, "-")

				if len(parts) == 1 {
					if !isDigit(parts[0]) {
						// Abort expansion on invalid input.
						return result, 0
					}
					result = append(result, prefix+parts[0]+suffix)
				} else if len(parts) == 2 {
					if !isDigit(parts[0]) || !isDigit(parts[1]) {
						return result, 0
					}
					sigFigures := len(parts[0])
					startNum := toInt(parts[0])
					endNum := toInt(parts[1])
					for num := startNum; num <= endNum; num++ {
						result = append(result, fmt.Sprintf("%s%0*d%s", prefix, sigFigures, num, suffix))
					}
				}
			}
		} else {
			// No brackets; keep the string as is.
			result = append(result, host)
		}
	}

	return result, expansionCount
}

// splitTopLevel splits s on commas that are not inside square brackets.
func splitTopLevel(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// toInt converts a numeric string to an integer. Assumes valid input.
func toInt(s string) int {
	num, _ := strconv.Atoi(s)
	return num
}

// isDigit returns true if s consists solely of ASCII digits.
func isDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
