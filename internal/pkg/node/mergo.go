package node

import (
	"net"
	"reflect"
	"strings"

	"dario.cat/mergo"
	"github.com/mohae/deepcopy"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// getNodeProfiles returns a deduplicated list of profile identifiers associated with the given node.
// It expands nested profiles from the node's profile list and then cleans the result by removing
// duplicate entries and any profiles prefixed with '~' (which denote negated profiles).
//
// If the specified node is not found in the configuration, the function returns an empty slice.
//
// Parameters:
//
//	id - the identifier of the node whose profiles are to be retrieved.
//
// Returns:
//
//	A slice of profile IDs associated with the given node, with nested profiles expanded and cleaned.
func (config *NodesYaml) getNodeProfiles(id string) (profiles []string) {
	visited := make(map[string]bool)

	if node, ok := config.Nodes[id]; ok {
		profiles = cleanList(config.getProfilesProfiles(node.Profiles, visited))
	}
	return profiles
}

// getProfilesProfiles recursively expands a list of profile identifiers by processing their nested profiles.
// For each profile ID in the input slice:
//   - If the profile ID begins with '~', it is treated as a negation and appended directly to the output.
//   - If the profile ID has already been processed (as tracked by the visited map), it is skipped to prevent duplication and infinite recursion.
//   - Otherwise, the profile ID is marked as visited. If a corresponding profile exists in the configuration's NodeProfiles,
//     its nested profiles are recursively processed and appended to the output before appending the profile ID itself.
//
// Parameters:
//
//	input   - A slice of profile IDs to be expanded.
//	visited - A map tracking profile IDs that have already been processed to avoid duplicates and cycles.
//
// Returns:
//
//	A slice of profile IDs that includes the original and all recursively expanded nested profiles.
func (config *NodesYaml) getProfilesProfiles(input []string, visited map[string]bool) (output []string) {
	for _, id := range input {
		if strings.HasPrefix(id, "~") {
			output = append(output, id)
			continue
		} else if visited[id] {
			continue
		} else {
			visited[id] = true
			if profile, ok := config.NodeProfiles[id]; ok {
				output = append(output, config.getProfilesProfiles(profile.Profiles, visited)...)
			}
			output = append(output, id)
		}
	}
	return output
}

type Transformer struct{}

func (t Transformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(net.IP{}) {
		return func(dst, src reflect.Value) error {
			if !src.IsValid() || !src.CanSet() {
				return nil
			}
			dst.Set(src)
			return nil
		}
	} else if typ.Kind() == reflect.Interface {
		return func(dst, src reflect.Value) error {
			if !src.IsValid() || src.IsZero() {
				return nil
			}

			// Handle merging of concrete values
			switch src.Interface().(type) {
			case map[string]interface{}:
				if dst.IsNil() {
					dst.Set(reflect.New(src.Elem().Type()).Elem())
				}
				return mergo.Merge(dst.Interface(), src.Interface(), mergo.WithAppendSlice, mergo.WithOverride, mergo.WithTransformers(t))
			case []interface{}:
				dst.Set(src)
			default:
				dst.Set(src)
			}
			return nil
		}
	}
	return nil
}

// MergeNode merges the configuration of a node identified by `id` with all the profiles
// associated with it, producing a fully composed `Node` and a `fieldMap` detailing the
// sources of various configuration fields.
//
// It works by:
//   - Retrieving the base node configuration using `GetNodeOnly`.
//   - Gathering all profile IDs associated with the node via `getNodeProfiles`.
//   - For each profile:
//   - Merging fields from a deep copy of each profile into the node,
//     recording the origin of each configuration field (i.e., which profile provided it)
//     in a `fieldMap` so that traceability is maintained.
//   - Finally, merging the original node configuration back into the processed node, ensuring
//     that any fields not set by the profiles are preserved, and updating the `fieldMap`
//     accordingly.
//
// Parameters:
// - id: The identifier of the node to be merged with its profiles.
//
// Returns:
// - node: The resulting merged `Node` configuration.
// - fields: A `fieldMap` detailing the source(s) of each configuration field.
// - err: An error if any node or profile retrieval or merging operations fail.
func (config *NodesYaml) MergeNode(id string) (node Node, fields fieldMap, err error) {
	node, err = config.GetNodeOnly(id)
	if err != nil {
		return node, fields, err
	}
	originalNode := node
	node = Node{}

	fields = make(fieldMap)

	for _, profileID := range config.getNodeProfiles(id) {
		if profile, err := config.GetProfile(profileID); err != nil {
			wwlog.Warn("profile not found: %s", profileID)
			continue
		} else {
			profile := deepcopy.Copy(profile)
			if err = merge(&node.Profile, profile, fields, profileID, profileID); err != nil {
				return node, fields, err
			}
		}
	}

	if err = merge(&node, originalNode, fields, "", id); err != nil {
		return node, fields, err
	}

	node.Profiles = originalNode.Profiles
	if len(node.Profiles) > 0 {
		fields.Set("Profiles", "", strings.Join(originalNode.Profiles, ","))
		fields["Profiles"].Source = ""
	} else {
		delete(fields, "Profiles")
	}

	node.id = id
	node.valid = true
	node.updatePrimaryNetDev()
	node.Profile.cleanLists()
	return node, fields, nil
}

// merge merges the fields of src (a data object) into dst (a pointer) associated with it. Used by
// MergeNode to provide consistent behavior when merging profiles and nodes.
//
// merge further tracks the source of each field in the provided fields.
//
// Because the source label behavior differs between multi-valued fields (e.g., slices) and
// single-valued slices, two source names must be provided: srcName is used for single-valued
// fields, and multipleSrcName is used for multi-sourced fields.
//
// Returns an error if the merging operation fails.
func merge(dest, src interface{}, fields fieldMap, srcName string, multipleSrcName string) error {
	if err := mergo.Merge(dest, src, mergo.WithAppendSlice, mergo.WithOverride, mergo.WithTransformers(Transformer{})); err != nil {
		return err
	}

	for _, fieldName := range listFields(src) {
		if value, err := getNestedFieldValue(src, fieldName); err == nil && valueStr(value) != "" {
			srcName := srcName
			prevSource := fields.Source(fieldName)
			if prevSource != "" {
				switch value.Kind() {
				case reflect.Slice:
					if value.Type() != reflect.TypeOf(net.IP{}) {
						srcName = strings.Join([]string{prevSource, multipleSrcName}, ",")
					}
				case reflect.Interface:
					if _, ok := value.Interface().([]interface{}); ok {
						srcName = strings.Join([]string{prevSource, multipleSrcName}, ",")
					}
				}
			}
			if value, err := getNestedFieldString(reflect.ValueOf(dest).Elem().Interface(), fieldName); err == nil {
				fields.Set(fieldName, srcName, value)
			}
		}
	}
	return nil
}
