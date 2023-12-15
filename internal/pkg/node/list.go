package node

import (
	"reflect"
)

/*
struct to hold the fields of GetFields
*/
type NodeFields struct {
	Field  string
	Source string
	Value  string
}

/*
Get all the info out of NodeConf. If emptyFields is set true, all fields are shown not only the ones with effective values
*/
func (nodeYml *NodeYaml) GetFields(node NodeConf, emptyFields bool) (output []NodeFields) {
	fieldMap := make(map[string]NodeFields)
	for _, p := range node.Profiles {
		if profile, ok := nodeYml.NodeProfiles[p]; ok {
			recursiveFields(profile, emptyFields, "", fieldMap, p)
		}
	}
	recursiveFields(node, emptyFields, "", fieldMap, "")
	return output
}

/*
Internal function which travels through all fields of a NodeConf and for this
reason needs tb called via interface{}
*/
func recursiveFields(obj interface{}, emptyFields bool, prefix string,
	fieldMap map[string]NodeFields, source string) {
	valObj := reflect.ValueOf(obj)
	typeObj := reflect.TypeOf(obj)
	for i := 0; i < typeObj.Elem().NumField(); i++ {
		if valObj.Elem().Field(i).IsValid() {
			if valObj.Elem().Field(i).String() != "" {
				fieldMap[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name,
					Source: source,
					Value:  valObj.Elem().Field(i).String(),
				}
			} else if emptyFields {
				fieldMap[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name + "[]",
					Source: source,
				}
			}
		} else if typeObj.Elem().Field(i).Type.Kind() == reflect.Map {
			mapIter := valObj.Elem().Field(i).MapRange()
			for mapIter.Next() {
				recursiveFields(mapIter.Value().Interface(),
					emptyFields, prefix+typeObj.Elem().Field(i).Name+"["+mapIter.Key().String()+"].", fieldMap, source)
			}
			if valObj.Elem().Field(i).Len() == 0 && emptyFields {
				fieldMap[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
					Field: prefix + typeObj.Elem().Field(i).Name + "[]",
				}
			}
		} else if typeObj.Elem().Field(i).Type.Kind() == reflect.Ptr {
			recursiveFields(valObj.Elem().Field(i).Interface(), emptyFields, prefix+typeObj.Elem().Field(i).Name+".", fieldMap, source)
		}
	}
}
