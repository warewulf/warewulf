package node

import (
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
func (nodeYml *NodeYaml) GetFields(node NodeConf) (output []NodeFields) {
	nodeMap := make(fieldMap)
	for _, p := range node.Profiles {
		if profile, ok := nodeYml.nodeProfiles[p]; ok {
			nodeMap.recursiveFields(profile, "", p)
		}
	}
	rawNode, _ := nodeYml.GetNodeOnlyPtr(node.id)
	nodeMap.recursiveFields(rawNode, "", "")
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
func (nodeYml *NodeYaml) GetFieldsProfile(profile ProfileConf) (output []NodeFields) {
	profileMap := make(fieldMap)
	profileMap.recursiveFields(&profile, "", "")
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
func (fieldMap fieldMap) recursiveFields(obj interface{}, prefix string, source string) {
	valObj := reflect.ValueOf(obj)
	typeObj := reflect.TypeOf(obj)
	if valObj.IsNil() {
		return
	}
	for i := 0; i < typeObj.Elem().NumField(); i++ {
		if valObj.Elem().Field(i).IsValid() {
			if !typeObj.Elem().Field(i).IsExported() {
				continue
			}
			switch typeObj.Elem().Field(i).Type.Kind() {
			case reflect.Map:
				mapIter := valObj.Elem().Field(i).MapRange()
				for mapIter.Next() {
					if mapIter.Value().Kind() == reflect.String {
						fieldMap[prefix+typeObj.Elem().Field(i).Name+"["+mapIter.Key().String()+"]"] = &NodeFields{
							Field:  prefix + typeObj.Elem().Field(i).Name + "[" + mapIter.Key().String() + "]",
							Source: source,
							Value:  mapIter.Value().String(),
						}
					} else {
						fieldMap.recursiveFields(mapIter.Value().Interface(), prefix+typeObj.Elem().Field(i).Name+"["+mapIter.Key().String()+"].", source)
					}
				}
				if valObj.Elem().Field(i).Len() == 0 {
					fieldMap[prefix+typeObj.Elem().Field(i).Name] = &NodeFields{
						Field: prefix + typeObj.Elem().Field(i).Name + "[]",
					}
				}
			case reflect.Struct:
				fieldMap.recursiveFields(valObj.Elem().Field(i).Addr().Interface(), "", source)
			case reflect.Ptr:
				if valObj.Elem().Field(i).Addr().IsValid() {
					fieldMap.recursiveFields(valObj.Elem().Field(i).Interface(), prefix+typeObj.Elem().Field(i).Name+".", source)
				}
			default:
				if _, ok := fieldMap[prefix+typeObj.Elem().Field(i).Name]; !ok {
					fieldMap[prefix+typeObj.Elem().Field(i).Name] = &NodeFields{
						Field:  prefix + typeObj.Elem().Field(i).Name,
						Source: source,
					}
				}

				switch typeObj.Elem().Field(i).Type {
				case reflect.TypeOf([]string{}):
					vals := (valObj.Elem().Field(i).Interface()).([]string)
					src_str := source
					if oldVal, ok := fieldMap[prefix+typeObj.Elem().Field(i).Name]; ok {
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
					fieldMap[prefix+typeObj.Elem().Field(i).Name] = &NodeFields{
						Field:  prefix + typeObj.Elem().Field(i).Name,
						Source: src_str,
						Value:  strings.Join(vals, ","),
					}
				case reflect.TypeOf(net.IP{}):
					val := (valObj.Elem().Field(i).Interface()).(net.IP)
					if val != nil {
						fieldMap[prefix+typeObj.Elem().Field(i).Name].Set(source, val.String())
					}
				case reflect.TypeOf(true):
					val := (valObj.Elem().Field(i).Interface()).(bool)
					if val {
						fieldMap[prefix+typeObj.Elem().Field(i).Name].Set(source, strconv.FormatBool(val))
					}
				default:
					fieldMap[prefix+typeObj.Elem().Field(i).Name].Set(source, valObj.Elem().Field(i).String())
				}

			}
		}
	}
}
