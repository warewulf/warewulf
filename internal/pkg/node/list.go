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
Get all the info out of NodeInfo. If emptyFields is set true, all fields
are shown not only the ones with effective values
*/
func (node *NodeInfo) GetFields(emptyFields bool) (output []NodeFields) {
	return recursiveFields(node, emptyFields, "")
}

/*
Internal function which travels through all fields of a NodeInfo and for this
reason needs tb called via interface{}
*/
func recursiveFields(obj interface{}, emptyFields bool, prefix string) (output []NodeFields) {
	valObj := reflect.ValueOf(obj)
	typeObj := reflect.TypeOf(obj)
	for i := 0; i < typeObj.Elem().NumField(); i++ {
		if typeObj.Elem().Field(i).Type == reflect.TypeOf(Entry{}) {
			myField := valObj.Elem().Field(i).Interface().(Entry)
			if emptyFields || myField.Get() != "" {
				output = append(output, NodeFields{
					Field:  prefix + typeObj.Elem().Field(i).Name,
					Source: myField.Source(),
					Value:  myField.Print(),
				})
			}
		} else if typeObj.Elem().Field(i).Type == reflect.TypeOf(map[string]*Entry{}) {
			for key, val := range valObj.Elem().Field(i).Interface().(map[string]*Entry) {
				if emptyFields || val.Get() != "" {
					output = append(output, NodeFields{
						Field:  prefix + typeObj.Elem().Field(i).Name + "[" + key + "]",
						Source: val.Source(),
						Value:  val.Print(),
					})
				}
			}
			if valObj.Elem().Field(i).Len() == 0 && emptyFields {
				output = append(output, NodeFields{
					Field: prefix + typeObj.Elem().Field(i).Name + "[]",
				})
			}
		} else if typeObj.Elem().Field(i).Type.Kind() == reflect.Map {
			mapIter := valObj.Elem().Field(i).MapRange()
			for mapIter.Next() {
				nestedOut := recursiveFields(mapIter.Value().Interface(), emptyFields, prefix+typeObj.Elem().Field(i).Name+"["+mapIter.Key().String()+"].")
				if len(nestedOut) == 0 {
					output = append(output, NodeFields{
						Field: prefix + typeObj.Elem().Field(i).Name + "[" + mapIter.Key().String() + "]",
					})
				} else {
					output = append(output, nestedOut...)
				}
			}
			if valObj.Elem().Field(i).Len() == 0 && emptyFields {
				output = append(output, NodeFields{
					Field: prefix + typeObj.Elem().Field(i).Name + "[]",
				})
			}
		} else if typeObj.Elem().Field(i).Type.Kind() == reflect.Ptr {
			nestedOut := recursiveFields(valObj.Elem().Field(i).Interface(), emptyFields, prefix+typeObj.Elem().Field(i).Name+".")
			output = append(output, nestedOut...)
		}
	}
	return
}
