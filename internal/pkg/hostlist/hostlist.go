package hostlist

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const Docstring = "Node patterns are a comma-separated list of individual patterns.\nEach pattern can either be a full node name or a node range like node[01-03,05]."

// Expand takes a slice of host strings, possibly containing comma-separated
// values and bracketed ranges (e.g. "node[01-03]") and returns a fully expanded
// slice of host names.
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

	return expanded
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

// Compress is the inverse of Expand: it returns a hostlist-style string in
// which numeric suffixes sharing a prefix and zero-pad width are collapsed
// into bracket notation. E.g. ["n01","n02","n03","n05"] -> "n[01-03,05]".
// Names without a trailing digit run are emitted as-is.
func Compress(ids []string) string {
	type group struct {
		prefix   string
		width    int
		nums     []int
		single   string
		firstIdx int
	}
	groups := map[string]*group{}
	var order []string
	for idx, id := range ids {
		i := len(id)
		for i > 0 && id[i-1] >= '0' && id[i-1] <= '9' {
			i--
		}
		var key string
		if i == len(id) {
			key = "\x00" + id
		} else {
			key = id[:i] + "\x00" + strconv.Itoa(len(id)-i)
		}
		g, ok := groups[key]
		if !ok {
			g = &group{firstIdx: idx}
			if i == len(id) {
				g.single = id
			} else {
				g.prefix = id[:i]
				g.width = len(id) - i
			}
			groups[key] = g
			order = append(order, key)
		}
		if g.width > 0 {
			n, _ := strconv.Atoi(id[i:])
			g.nums = append(g.nums, n)
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		return groups[order[i]].firstIdx < groups[order[j]].firstIdx
	})

	var parts []string
	for _, k := range order {
		g := groups[k]
		if g.width == 0 {
			parts = append(parts, g.single)
			continue
		}
		sort.Ints(g.nums)
		if len(g.nums) == 1 {
			parts = append(parts, fmt.Sprintf("%s%0*d", g.prefix, g.width, g.nums[0]))
			continue
		}
		var ranges []string
		start, prev := g.nums[0], g.nums[0]
		flush := func() {
			if start == prev {
				ranges = append(ranges, fmt.Sprintf("%0*d", g.width, start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%0*d-%0*d", g.width, start, g.width, prev))
			}
		}
		for _, v := range g.nums[1:] {
			if v == prev+1 {
				prev = v
				continue
			}
			flush()
			start, prev = v, v
		}
		flush()
		parts = append(parts, fmt.Sprintf("%s[%s]", g.prefix, strings.Join(ranges, ",")))
	}
	return strings.Join(parts, ",")
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
