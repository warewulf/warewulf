package overlaydiff

import "strings"

// FilterOptions controls change filtering for preview and interactive selection.
type FilterOptions struct {
	Only       []ChangeType
	PathPrefix []string
}

// ParseChangeTypes validates and parses repeatable --only values.
func ParseChangeTypes(values []string) ([]ChangeType, error) {
	if len(values) == 0 {
		return nil, nil
	}

	result := make([]ChangeType, 0, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		switch ChangeType(normalized) {
		case ChangeAdded, ChangeModified, ChangeModeChanged:
			result = append(result, ChangeType(normalized))
		default:
			return nil, fmtInvalidChangeFilter(value)
		}
	}
	return result, nil
}

// fmtInvalidChangeFilter wraps an invalid --only value in a typed error.
func fmtInvalidChangeFilter(value string) error {
	return &ErrInvalidFilter{Value: value}
}

// ErrInvalidFilter is returned for invalid --only filter values.
type ErrInvalidFilter struct {
	Value string
}

// Error returns a user-facing message for invalid --only filter values.
func (e *ErrInvalidFilter) Error() string {
	return "invalid --only value \"" + e.Value + "\": expected added|modified|mode-changed"
}

// FilterChanges returns a new slice containing only entries matching the filter.
func FilterChanges(changes []Change, options FilterOptions) []Change {
	if len(changes) == 0 {
		return nil
	}

	onlySet := make(map[ChangeType]struct{}, len(options.Only))
	for _, value := range options.Only {
		onlySet[value] = struct{}{}
	}

	prefixes := NormalizeExcludes(options.PathPrefix)

	result := make([]Change, 0, len(changes))
	for _, change := range changes {
		if len(onlySet) > 0 {
			if _, ok := onlySet[change.Change]; !ok {
				continue
			}
		}

		if len(prefixes) > 0 {
			matches := false
			for _, prefix := range prefixes {
				// Match exact path or children under that path, but not sibling prefixes.
				if change.Path == prefix || strings.HasPrefix(change.Path, prefix+"/") {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}

		result = append(result, change)
	}

	return result
}
