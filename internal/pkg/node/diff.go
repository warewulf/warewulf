package node

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Change is a single field-level difference between two profiles.
type Change struct {
	Path   string
	Before string
	After  string
}

// Clone returns a deep copy of the profile via YAML round-trip.
func (p *Profile) Clone() *Profile {
	data, err := yaml.Marshal(p)
	if err != nil {
		return nil
	}
	out := NewProfile(p.Id())
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil
	}
	return &out
}

// Clone returns a deep copy of the node via YAML round-trip.
func (n *Node) Clone() *Node {
	data, err := yaml.Marshal(n)
	if err != nil {
		return nil
	}
	out := NewNode(n.Id())
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil
	}
	return &out
}

// DiffProfile returns the field-level differences between before and after.
// Paths use lopt-tag names when available (falling back to lowercase field
// names) and bracket notation for map keys (e.g. "tags[hostname]",
// "netdev[default].ipaddr"). The result is sorted by path.
func DiffProfile(before, after *Profile) []Change {
	var changes []Change
	diffStruct(reflect.ValueOf(before).Elem(), reflect.ValueOf(after).Elem(), "", &changes)
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}

func diffStruct(before, after reflect.Value, prefix string, out *[]Change) {
	t := before.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Tag.Get("lopt")
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if f.Anonymous {
			path = prefix
		}
		diffValue(before.Field(i), after.Field(i), path, out)
	}
}

func diffValue(before, after reflect.Value, path string, out *[]Change) {
	switch before.Kind() {
	case reflect.Struct:
		diffStruct(before, after, path, out)
	case reflect.Ptr:
		bNil := !before.IsValid() || before.IsNil()
		aNil := !after.IsValid() || after.IsNil()
		if bNil && aNil {
			return
		}
		var b, a reflect.Value
		if bNil {
			b = reflect.New(before.Type().Elem()).Elem()
		} else {
			b = before.Elem()
		}
		if aNil {
			a = reflect.New(after.Type().Elem()).Elem()
		} else {
			a = after.Elem()
		}
		if b.Kind() == reflect.Struct {
			diffStruct(b, a, path, out)
		} else if bs, as := fmtScalar(b), fmtScalar(a); bs != as {
			*out = append(*out, Change{Path: path, Before: bs, After: as})
		}
	case reflect.Map:
		diffMap(before, after, path, out)
	case reflect.Slice:
		// Byte slices (e.g. net.IP) render via their Stringer rather than as a list of bytes.
		if before.Type().Elem().Kind() == reflect.Uint8 {
			bs, as := fmtScalar(before), fmtScalar(after)
			if bs != as {
				*out = append(*out, Change{Path: path, Before: bs, After: as})
			}
			return
		}
		bs, as := fmtSlice(before), fmtSlice(after)
		if bs != as {
			*out = append(*out, Change{Path: path, Before: bs, After: as})
		}
	default:
		b, a := fmtScalar(before), fmtScalar(after)
		if b != a {
			*out = append(*out, Change{Path: path, Before: b, After: a})
		}
	}
}

func diffMap(before, after reflect.Value, prefix string, out *[]Change) {
	keys := make(map[string]bool)
	collect := func(v reflect.Value) {
		if !v.IsValid() || v.IsNil() {
			return
		}
		for _, k := range v.MapKeys() {
			keys[k.String()] = true
		}
	}
	collect(before)
	collect(after)
	ordered := make([]string, 0, len(keys))
	for k := range keys {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)
	for _, k := range ordered {
		bv := mapGet(before, k)
		av := mapGet(after, k)
		path := fmt.Sprintf("%s[%s]", prefix, k)
		diffValue(bv, av, path, out)
	}
}

func mapGet(m reflect.Value, key string) reflect.Value {
	if !m.IsValid() || m.IsNil() {
		// Return a zero value of the element type so callers can recurse safely.
		return reflect.Zero(m.Type().Elem())
	}
	v := m.MapIndex(reflect.ValueOf(key))
	if !v.IsValid() {
		return reflect.Zero(m.Type().Elem())
	}
	return v
}

func fmtScalar(v reflect.Value) string {
	if !v.IsValid() {
		return "<unset>"
	}
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return "<unset>"
		}
		return fmt.Sprintf("%q", v.String())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Interface:
		if v.IsNil() {
			return "<unset>"
		}
		return fmt.Sprintf("%v", v.Interface())
	}
	if v.CanInterface() {
		iv := v.Interface()
		if s, ok := iv.(fmt.Stringer); ok {
			out := s.String()
			if out == "" || out == "<nil>" {
				return "<unset>"
			}
			return out
		}
		return fmt.Sprintf("%v", iv)
	}
	return "<unset>"
}

// FormatChanges renders a per-entity change map as a multi-line summary.
// Entities (nodes or profiles) with the same change-set are grouped together
// onto a single header line. Returns "" when there are no changes.
func FormatChanges(entityChanges map[string][]Change) string {
	type group struct {
		ids     []string
		changes []Change
	}
	groups := map[string]*group{}
	var order []string
	for id, ch := range entityChanges {
		if len(ch) == 0 {
			continue
		}
		key := fingerprint(ch)
		if g, ok := groups[key]; ok {
			g.ids = append(g.ids, id)
		} else {
			groups[key] = &group{ids: []string{id}, changes: ch}
			order = append(order, key)
		}
	}
	if len(groups) == 0 {
		return ""
	}
	sort.Slice(order, func(i, j int) bool {
		return groups[order[i]].ids[0] < groups[order[j]].ids[0]
	})
	var b strings.Builder
	for i, key := range order {
		g := groups[key]
		sort.Strings(g.ids)
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s:\n", strings.Join(g.ids, ", "))
		for _, c := range g.changes {
			fmt.Fprintf(&b, "  %s: %s → %s\n", c.Path, c.Before, c.After)
		}
	}
	return b.String()
}

func fingerprint(changes []Change) string {
	parts := make([]string, len(changes))
	for i, c := range changes {
		parts[i] = c.Path + "\x00" + c.Before + "\x00" + c.After
	}
	return strings.Join(parts, "\n")
}

func fmtSlice(v reflect.Value) string {
	if !v.IsValid() || v.IsNil() || v.Len() == 0 {
		return "[]"
	}
	parts := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		parts[i] = fmt.Sprintf("%v", v.Index(i).Interface())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
