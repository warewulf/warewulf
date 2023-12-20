package node

import (
	"fmt"
	"net"
	"reflect"
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

type fieldMap map[string]NodeFields

/*
Get all the info out of NodeConf. If emptyFields is set true, all fields are shown not only the ones with effective values
*/
func (nodeYml *NodeYaml) GetFields(node NodeConf, emptyFields bool) (output []NodeFields) {
	nodeMap := make(fieldMap)
	for _, p := range node.Profiles {
		if profile, ok := nodeYml.nodeProfiles[p]; ok {
			nodeMap.recursiveFields(profile, emptyFields, "", p)
		}
	}
	nodeMap.recursiveFields(&node, emptyFields, "", "")
	for _, elem := range nodeMap {
		output = append(output, elem)
	}
	return output
}

/*
Get all the info out of ProfileConf. If emptyFields is set true, all fields are shown not only the ones with effective values
*/
func (nodeYml *NodeYaml) GetFieldsProfile(profile ProfileConf, emptyFields bool) (output []NodeFields) {
	profileMap := make(fieldMap)
	profileMap.recursiveFields(&profile, emptyFields, "", "")
	for _, elem := range profileMap {
		output = append(output, elem)
	}
	return output
}

/*
Internal function which travels through all fields of a NodeConf and for this
reason needs to be called via interface{}
*/
func (fieldMap *fieldMap) recursiveFields(obj interface{}, emptyFields bool, prefix string, source string) {
	valObj := reflect.ValueOf(obj)
	typeObj := reflect.TypeOf(obj)
	for i := 0; i < typeObj.Elem().NumField(); i++ {
		fmt.Printf("name: %s\n", typeObj.Elem().Field(i).Name)
		if valObj.Elem().Field(i).IsValid() {
			if !typeObj.Elem().Field(i).IsExported() {
				continue
			}
			if valObj.Elem().Field(i).Kind() == reflect.String && valObj.Elem().Field(i).String() != "" {
				fmt.Printf("string: %s\n", valObj.Elem().Field(i).String())
				(*fieldMap)[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name,
					Source: source,
					Value:  valObj.Elem().Field(i).String(),
				}
			} else if emptyFields {
				(*fieldMap)[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name + "[]",
					Source: source,
				}
			} else if typeObj.Elem().Field(i).Type == reflect.TypeOf([]string{}) && valObj.Elem().Field(i).Len() != 0 {
				vals := (valObj.Elem().Field(i).Interface()).([]string)
				(*fieldMap)[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name,
					Source: source,
					Value:  strings.Join(vals, ","),
				}
			} else if typeObj.Elem().Field(i).Type == reflect.TypeOf(net.IP{}) {
				val := (valObj.Elem().Field(i).Interface()).(net.IP)
				(*fieldMap)[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name,
					Source: source,
					Value:  val.String(),
				}
			} else if typeObj.Elem().Field(i).Type.Kind() == reflect.Map {
				mapIter := valObj.Elem().Field(i).MapRange()
				for mapIter.Next() {
					fieldMap.recursiveFields(mapIter.Value().Interface(),
						emptyFields, prefix+typeObj.Elem().Field(i).Name+"["+mapIter.Key().String()+"].", source)
				}
				if valObj.Elem().Field(i).Len() == 0 && emptyFields {
					(*fieldMap)[prefix+typeObj.Elem().Field(i).Name] = NodeFields{
						Field: prefix + typeObj.Elem().Field(i).Name + "[]",
					}
				}

			} /*else if typeObj.Elem().Field(i).Type.Kind() == reflect.Ptr {
				fieldMap.recursiveFields(valObj.Elem().Field(i).Interface(), emptyFields, prefix+typeObj.Elem().Field(i).Name+".", source)
			}*/
		}
	}
}
