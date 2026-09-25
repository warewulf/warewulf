package config

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// An OverlayList is a comma-separated list of overlays that a service
// applies when it is configured. Overlays are applied left-to-right,
// with files from the rightmost overlay taking precedence.
type OverlayList []string

// UnmarshalYAML accepts either a comma-separated string or a yaml
// sequence of overlay names.
func (overlays *OverlayList) UnmarshalYAML(node *yaml.Node) error {
	var commaSeparated string
	if err := node.Decode(&commaSeparated); err == nil {
		*overlays = nil
		for _, name := range strings.Split(commaSeparated, ",") {
			if name = strings.TrimSpace(name); name != "" {
				*overlays = append(*overlays, name)
			}
		}
		return nil
	}
	var names []string
	if err := node.Decode(&names); err != nil {
		return err
	}
	*overlays = names
	return nil
}

// MarshalYAML writes the overlay list as a comma-separated string.
func (overlays OverlayList) MarshalYAML() (interface{}, error) {
	return strings.Join(overlays, ","), nil
}
