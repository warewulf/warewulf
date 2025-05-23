// Package systemd provides functions to escape arbitrary strings into valid
// systemd unit names, including path-specific escaping.
package systemd

import (
	"fmt"
	"strings"
)

// Escape applies systemd escaping to an arbitrary string (non-path).
func Escape(s string) string {
	return escape(s, false)
}

// EscapePath applies systemd escaping to a filesystem path.
// Special cases:
//
//	"/"    -> "-"
//	"/foo//bar/baz/" -> "foo-bar-baz"
func EscapePath(s string) string {
	return escape(s, true)
}

// escape is the core implementation. If path=true, it first strips
// leading, trailing, and duplicate slashes, and encodes "/" as "-".
func escape(input string, path bool) string {
	// Handle path normalization
	if path {
		// root -> single dash
		if input == "/" {
			return "-"
		}
		// split out segments (drops empty segments)
		parts := strings.FieldsFunc(input, func(r rune) bool { return r == '/' })
		input = strings.Join(parts, "/")
	}

	var sb strings.Builder
	first := true

	for i := 0; i < len(input); i++ {
		b := input[i]
		switch {
		case b == '/':
			// path-mode slash -> dash
			sb.WriteByte('-')
		case isAlnum(b) || b == ':' || b == '_':
			// safe characters
			sb.WriteByte(b)
		case b == '.' && !first:
			// dot allowed except at start
			sb.WriteByte(b)
		default:
			// C-style hex escape
			sb.WriteString(fmt.Sprintf("\\x%02x", b))
		}
		first = false
	}

	return sb.String()
}

// isAlnum returns true if b is ASCII letter or digit.
func isAlnum(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}
