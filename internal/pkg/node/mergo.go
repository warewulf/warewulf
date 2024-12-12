package node

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"strings"

	"dario.cat/mergo"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func copyProfile(this Profile) (Profile, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	profile := Profile{}
	if err := enc.Encode(this); err != nil {
		return profile, err
	} else {
		if err := dec.Decode(&profile); err != nil {
			return profile, err
		} else {
			return profile, nil
		}
	}
}

func (config *NodesYaml) MergeNode(id string) (node Node, fields fieldMap, err error) {
	node, err = config.GetNodeOnly(id)
	if err != nil {
		return node, fields, err
	}
	originalNode := node
	node = EmptyNode()

	fields = make(fieldMap)

	for _, profileID := range cleanList(originalNode.Profiles) {
		if profile, err := config.GetProfile(profileID); err != nil {
			wwlog.Warn("profile not found: %s", profileID)
			continue
		} else if profile, err := copyProfile(profile); err != nil {
			wwlog.Warn("error processing profile %s: %v", profileID, err)
			continue
		} else {
			if err = mergo.Merge(&node.Profile, profile, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
				return node, fields, err
			}
			for _, fieldName := range listFields(profile) {
				if value, err := getNestedFieldValue(profile, fieldName); err == nil && valueStr(value) != "" {
					source := profileID
					prevSource := fields.Source(fieldName)
					if value.Kind() == reflect.Slice && prevSource != "" {
						source = strings.Join([]string{prevSource, source}, ",")
					}
					if value, err := getNestedFieldString(node, fieldName); err == nil {
						fields.Set(fieldName, source, value)
					}
				}
			}
		}
	}

	if err = mergo.Merge(&node, originalNode, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
		return node, fields, err
	}
	for _, fieldName := range listFields(originalNode) {
		if value, err := getNestedFieldValue(originalNode, fieldName); err == nil && valueStr(value) != "" {
			source := ""
			prevSource := fields.Source(fieldName)
			if value.Kind() == reflect.Slice && prevSource != "" {
				source = strings.Join([]string{prevSource, id}, ",")
			}
			if value, err := getNestedFieldString(node, fieldName); err == nil {
				fields.Set(fieldName, source, value)
			}
		}
	}

	node.id = id
	node.valid = true
	node.updatePrimaryNetDev()
	return node, fields, nil
}
