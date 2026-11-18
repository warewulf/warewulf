package config

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

// Routes that warewulf:secure can require a privileged source port for.
const (
	SecureRouteRuntime = "runtime"
	SecureRouteSystem  = "system"
	SecureRouteFiles   = "files"
)

// secureRouteNames lists every securable route, in the order a list is
// written back to warewulf.conf.
var secureRouteNames = []string{SecureRouteRuntime, SecureRouteSystem, SecureRouteFiles}

// secureTrueRoutes are the routes secured by `secure: true`.
var secureTrueRoutes = []string{SecureRouteRuntime, SecureRouteFiles}

type secureForm int

const (
	secureFormBool secureForm = iota
	secureFormAll
	secureFormList
)

// SecureRoutes is the set of routes that require a privileged source
// port. warewulf.conf sets it as `true`, `false`, `all`, or a list of
// route names, and MarshalYAML writes it back in the same form.
//
// A nil *SecureRoutes means `true`.
type SecureRoutes struct {
	form   secureForm
	value  bool
	routes map[string]bool
}

// NewSecureRoutes returns a SecureRoutes in list form containing routes.
func NewSecureRoutes(routes ...string) *SecureRoutes {
	set := make(map[string]bool)
	for _, route := range routes {
		set[route] = true
	}
	return &SecureRoutes{form: secureFormList, routes: set}
}

// Has reports whether route requires a privileged source port.
func (s *SecureRoutes) Has(route string) bool {
	if s == nil {
		return slices.Contains(secureTrueRoutes, route)
	}
	switch s.form {
	case secureFormAll:
		return slices.Contains(secureRouteNames, route)
	case secureFormList:
		return s.routes[route]
	default:
		return s.value && slices.Contains(secureTrueRoutes, route)
	}
}

// With returns s with route added (enabled) or removed. If that does not
// change the set, With returns s unchanged, keeping its form. Otherwise it
// returns a list.
func (s *SecureRoutes) With(route string, enabled bool) *SecureRoutes {
	if s.Has(route) == enabled {
		return s
	}
	var routes []string
	for _, name := range secureRouteNames {
		if (name == route && enabled) || (name != route && s.Has(name)) {
			routes = append(routes, name)
		}
	}
	return NewSecureRoutes(routes...)
}

// WarnDeprecated logs a warning for each deprecated setting in conf.
func (conf *WarewulfConf) WarnDeprecated() {
	if conf != nil && conf.SecureFilesP != nil {
		wwlog.Warn(`warewulf:secure files is deprecated; include or omit "files" in warewulf:secure instead`)
	}
}

func secureRoutesError(value string) error {
	return fmt.Errorf("invalid warewulf:secure value %s: must be true, false, all, or a list of: %s",
		value, strings.Join(secureRouteNames, ", "))
}

// UnmarshalYAML accepts a bool, `all`, or a sequence of route names.
func (s *SecureRoutes) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var value bool
		if err := node.Decode(&value); err == nil {
			*s = SecureRoutes{form: secureFormBool, value: value}
			return nil
		}
		if node.Value == "all" {
			*s = SecureRoutes{form: secureFormAll}
			return nil
		}
		return secureRoutesError(strconv.Quote(node.Value))
	case yaml.SequenceNode:
		// yaml.v3 drops null entries, so a length mismatch means one was
		// present.
		var routes []string
		if err := node.Decode(&routes); err != nil || len(routes) != len(node.Content) {
			return secureRoutesError("list at line " + strconv.Itoa(node.Line))
		}
		for _, route := range routes {
			if !slices.Contains(secureRouteNames, route) {
				return secureRoutesError("entry " + strconv.Quote(route))
			}
		}
		*s = *NewSecureRoutes(routes...)
		return nil
	default:
		return secureRoutesError("at line " + strconv.Itoa(node.Line))
	}
}

// UnmarshalText parses text as YAML. It lets the `default` struct tag set
// a SecureRoutes.
func (s *SecureRoutes) UnmarshalText(text []byte) error {
	return yaml.Unmarshal(text, s)
}

// MarshalYAML writes the value in the form it was read. A list is written
// in a stable order.
func (s SecureRoutes) MarshalYAML() (interface{}, error) {
	switch s.form {
	case secureFormAll:
		return "all", nil
	case secureFormList:
		routes := []string{}
		for _, name := range secureRouteNames {
			if s.routes[name] {
				routes = append(routes, name)
			}
		}
		return routes, nil
	default:
		return s.value, nil
	}
}
