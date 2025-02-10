package hostlist

import (
	"fmt"
	"strconv"
	"strings"
)

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
