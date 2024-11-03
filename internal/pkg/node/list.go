package node

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

/*
struct to hold the fields of GetFields
*/
type NodeFields struct {
	Field  string
	Source string
	Value  string
}

func (f *NodeFields) Set(src, val string) {
	if val == "" {
		return
	}
	if f.Value == "" {
		f.Value = val
		f.Source = src
	} else if f.Source != "" {
		f.Value = val
		if src == "" {
			f.Source = "SUPERSEDED"
		} else {
			f.Source = src
		}
	}

}

type fieldMap map[string]*NodeFields

/*
Get all the info out of NodeConf. If emptyFields is set true, all fields are shown not only the ones with effective values
*/
func (nodeYml *NodesYaml) GetFields(node Node) (output []NodeFields) {
	nodeMap := make(fieldMap)
	for _, p := range node.Profiles {
		if profile, ok := nodeYml.NodeProfiles[p]; ok {
			nodeMap.importFields(profile, "", p)
		}
	}
	rawNode, _ := nodeYml.GetNodeOnlyPtr(node.id)
	nodeMap.importFields(rawNode, "", "")
	for _, elem := range nodeMap {
		if elem.Value != "" {
			output = append(output, *elem)
		}
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].Field < output[j].Field
	})
	return output
}

/*
Get all the info out of ProfileConf. If emptyFields is set true, all fields are shown not only the ones with effective values
*/
func (nodeYml *NodesYaml) GetFieldsProfile(profile Profile) (output []NodeFields) {
	profileMap := make(fieldMap)
	profileMap.importFields(&profile, "", "")
	for _, elem := range profileMap {
		if elem.Value != "" {
			output = append(output, *elem)
		}
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].Field < output[j].Field
	})
	return output
}

/*
Internal function which travels through all fields of a NodeConf and for this
reason needs to be called via interface{}
*/
func (fieldMap fieldMap) importFields(obj interface{}, prefix string, source string) {
	objValue := reflect.ValueOf(obj)
	objType := reflect.TypeOf(obj)
	if objValue.IsNil() {
		return
	}
	for i := 0; i < objType.Elem().NumField(); i++ {
		fieldValue := objValue.Elem().Field(i)
		field := objType.Elem().Field(i)
		if !fieldValue.IsValid() || !field.IsExported() {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Map:
			mapIter := fieldValue.MapRange()
			for mapIter.Next() {
				if mapIter.Value().Kind() == reflect.String {
					key := fmt.Sprintf("%s%s[%s]", prefix, field.Name, mapIter.Key().String())
					fieldMap[key] = &NodeFields{
						Field:  key,
						Source: source,
						Value:  mapIter.Value().String(),
					}
				} else {
					newPrefix := fmt.Sprintf("%s%s[%s].", prefix, field.Name, mapIter.Key().String())
					fieldMap.importFields(mapIter.Value().Interface(), newPrefix, source)
				}
			}
			if fieldValue.Len() == 0 {
				key := fmt.Sprintf("%s%s[]", prefix, field.Name)
				fieldMap[key] = &NodeFields{
					Field: key,
				}
			}
		case reflect.Struct: // inherited fields from a sub-entity
			fieldMap.importFields(fieldValue.Addr().Interface(), "", source)
		case reflect.Ptr:
			if fieldValue.Addr().IsValid() {
				if field.Type.Elem().Kind() == reflect.Bool {
					fieldMap.importField(field, fieldValue, prefix, source)
				} else {
					newPrefix := fmt.Sprintf("%s%s.", prefix, field.Name)
					fieldMap.importFields(fieldValue.Interface(), newPrefix, source)
				}
			}
		default:
			fieldMap.importField(field, fieldValue, prefix, source)
		}
	}
}

func (fieldMap fieldMap) importField(field reflect.StructField, fieldValue reflect.Value, prefix string, source string) {
	key := prefix + field.Name
	if _, ok := fieldMap[key]; !ok {
		fieldMap[key] = &NodeFields{
			Field:  key,
			Source: source,
		}
	}

	if field.Type == reflect.TypeOf([]string{}) {
		vals := (fieldValue.Interface()).([]string)
		src_str := source
		if oldVal, ok := fieldMap[key]; ok {
			if oldVal.Value != "" {
				if len(vals) > 0 {
					src_str = oldVal.Source + "+"
				} else {
					src_str = oldVal.Source
				}
				vals = append(vals, oldVal.Value)
			} else {
				src_str = oldVal.Source
			}
		}
		fieldMap[key] = &NodeFields{
			Field:  key,
			Source: src_str,
			Value:  strings.Join(vals, ","),
		}
	} else if field.Type == reflect.TypeOf(net.IP{}) {
		val := (fieldValue.Interface()).(net.IP)
		if val != nil {
			fieldMap[key].Set(source, val.String())
		}
	} else if field.Type.Kind() == reflect.Bool {
		fieldMap[key].Set(source, strconv.FormatBool(fieldValue.Bool()))
	} else if field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Bool {
		if fieldValue.Elem().IsValid() {
			fieldMap[key].Set(source, strconv.FormatBool(fieldValue.Elem().Bool()))
		}
	} else {
		fieldMap[key].Set(source, fieldValue.String())
	}
}
