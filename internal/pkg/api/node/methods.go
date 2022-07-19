package apinode

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
)

/*
Get the all the fields of a NodeInfo, keys are the lopt of NodeConf.
func GetFields(n interface{}) map[string]*wwapiv1.NodeField {
	return getFieldsOf(n, node.NodeConf{})
}
*/

func GetFields(n interface{}) map[string]*wwapiv1.NodeField {
	nodeType := reflect.TypeOf(n)
	nodeVal := reflect.ValueOf(n)
	fieldMap := make(map[string]*wwapiv1.NodeField)
	for i := 0; i < nodeType.NumField(); i++ {
		switch nodeType.Field(i).Type {
		case reflect.TypeOf(node.Entry{}):
			var myField wwapiv1.NodeField
			entry := nodeVal.Field(i).Interface().(node.Entry)
			myField.Source = entry.Source()
			myField.Value = entry.Get()
			myField.Print = entry.Print()
			fieldMap[nodeType.Field(i).Name] = &myField
		case reflect.TypeOf([]string{}):
			var myField wwapiv1.NodeField
			entry := nodeVal.Field(i).Interface().([]string)
			if len(entry) == 0 {
				myField.Value = node.NoValue
				myField.Print = node.NoValue
				myField.Source = node.NoValue
			} else {
				myField.Value = strings.Join(entry, ",")
				myField.Print = strings.Join(entry, ",")
				myField.Source = node.NoValue
				fieldMap[nodeType.Field(i).Name] = &myField
			}
		case reflect.TypeOf((*node.KernelEntry)(nil)):
			entry := nodeVal.Field(i).Elem().Interface().(node.KernelEntry)
			kernelMap := GetFields(entry)
			for key, val := range kernelMap {
				fieldMap["KernelEntry:"+key] = val
			}
		case reflect.TypeOf((*node.IpmiEntry)(nil)):
			entry := nodeVal.Field(i).Elem().Interface().(node.IpmiEntry)
			kernelMap := GetFields(entry)
			for key, val := range kernelMap {
				fieldMap["IpmiEntry:"+key] = val
			}
		case reflect.TypeOf(map[string]*node.Entry(nil)):
			keyMap := nodeVal.Field(i).Interface().(map[string]*node.Entry)
			for key, entr := range keyMap {
				var myField wwapiv1.NodeField
				myField.Source = entr.Source()
				myField.Value = entr.Get()
				myField.Print = entr.Print()
				fieldMap["key:"+key] = &myField
			}
		case reflect.TypeOf(map[string]*node.NetDevEntry(nil)):
			netMap := nodeVal.Field(i).Interface().(map[string]*node.NetDevEntry)
			for net, netdev := range netMap {
				netMapEntr := GetFields(*netdev)
				for key, val := range netMapEntr {
					fieldMap["NetDevEntry:"+net+":"+key] = val
				}
			}
		default:
			fmt.Println(nodeType.Field(i).Type)
		}
	}
	return fieldMap
}
