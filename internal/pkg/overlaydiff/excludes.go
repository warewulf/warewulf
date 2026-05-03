package overlaydiff

import (
	"path/filepath"
	"strings"
)

// DefaultExcludes are common runtime/log/cache paths excluded from scanning.
var DefaultExcludes = []string{
	"/var/log",
	"/var/cache",
	"/var/tmp",
	"/tmp",
	"/run",
	"/var/run",
	"/dev",
	"/proc",
	"/sys",
}

// ResolveExcludes merges defaults with overrides and normalizes the result.
func ResolveExcludes(overrides []string) []string {
	merged := make([]string, 0, len(DefaultExcludes)+len(overrides))
	merged = append(merged, DefaultExcludes...)
	merged = append(merged, overrides...)
	return NormalizeExcludes(merged)
}

// NormalizeExcludes ensures excludes are absolute-style paths with slashes
// and de-duplicates them while preserving order.
func NormalizeExcludes(excludes []string) []string {
	seen := make(map[string]struct{}, len(excludes))
	normalized := make([]string, 0, len(excludes))
	for _, exclude := range excludes {
		value := strings.TrimSpace(exclude)
		if value == "" {
			continue
		}
		value = filepath.ToSlash(filepath.Clean(value))
		if value == "." {
			continue
		}
		if !strings.HasPrefix(value, "/") {
			value = "/" + value
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}
