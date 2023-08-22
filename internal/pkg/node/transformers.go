package node

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
Populates a NodeConf struct (the one which goes to disk) from a
NodeInfo (which just lives in memory), with the values from all
the underlying entries using GetReal, so just the explicit values
go do disk.
*/
func (nodeConf *NodeConf) GetRealFrom(nodeInfo NodeInfo) {
	recursiveGetter(&nodeInfo, nodeConf, (*Entry).GetReal, (*Entry).GetRealSlice)
}

/*
Populates a NodeConf struct from a NodeInfo, with the combined
values from the underlying entries using Get.
*/
func (nodeConf *NodeConf) GetFrom(nodeInfo NodeInfo) {
	recursiveGetter(&nodeInfo, nodeConf, (*Entry).Get, (*Entry).GetSlice)
}

/*
Abstract function which populates a NodeConf from the given NodeInfo
via getter functions. Calls recursive itself for nested structures.
Panics if the NodeConf has fields which are not type of string,[]string,map[string]*ptr
*/
func recursiveGetter(
	source, target interface{},
	getter func(*Entry) string,
	getterSlice func(*Entry) []string) {
	sourceValue := reflect.ValueOf(source)
	targetType := reflect.TypeOf(target)
	targetValue := reflect.ValueOf(target)
	if targetValue.Elem().Kind() == reflect.Struct && sourceValue.Elem().Kind() == reflect.Struct {
		for i := 0; i < targetType.Elem().NumField(); i++ {
			sourceValueMatched := sourceValue.Elem().FieldByName(targetType.Elem().Field(i).Name)
			if sourceValueMatched.IsValid() {
				if sourceValueMatched.Type() == reflect.TypeOf(Entry{}) {
					// get the fields which are part of the struct
					switch targetValue.Elem().Field(i).Type() {
					case reflect.TypeOf(""):
						newValue := (targetValue.Elem().Field(i).Addr().Interface()).(*string)
						source := sourceValueMatched.Interface().(Entry)
						*newValue = getter(&source)
					case reflect.TypeOf([]string{}):
						newValue := (targetValue.Elem().Field(i).Addr().Interface()).(*[]string)
						source := sourceValueMatched.Interface().(Entry)
						*newValue = getterSlice(&source)
					default:
						panic(fmt.Errorf("can't convert an Entry to %s", targetValue.Elem().Field(i).Type()))
					}
				} else if sourceValueMatched.Kind() == reflect.Ptr {
					// if we get a pointer, initialize if empty and then have a recursive call
					if targetValue.Elem().Field(i).IsZero() {
						targetValue.Elem().Field(i).Set(reflect.New(targetType.Elem().Field(i).Type.Elem()))
					}
					recursiveGetter(sourceValueMatched.Interface(), targetValue.Elem().Field(i).Interface(), getter, getterSlice)
				} else if sourceValueMatched.Type().Kind() == reflect.Map {
					if targetValue.Elem().Field(i).IsZero() {
						targetValue.Elem().Field(i).Set(reflect.MakeMap(targetType.Elem().Field(i).Type))
					}
					sourceIter := sourceValueMatched.MapRange()
					if sourceValueMatched.Type() == reflect.TypeOf(map[string]*Entry{}) {
						// go over a simple map with strings
						for sourceIter.Next() {
							if !targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).IsValid() {
								str := getter((sourceIter.Value().Interface()).(*Entry))
								targetValue.Elem().Field(i).SetMapIndex(sourceIter.Key(), reflect.ValueOf(str))
							}
						}
					} else {
						// now the complicated map which contains pointers to objects
						for sourceIter.Next() {
							if !targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).IsValid() {
								newPtr := reflect.New(targetType.Elem().Field(i).Type.Elem().Elem())
								targetValue.Elem().Field(i).SetMapIndex(sourceIter.Key(), newPtr)
							}
							recursiveGetter(sourceIter.Value().Interface(), targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).Interface(), getter, getterSlice)

						}
					}
				}
			}
		}
	}
}

/*
Populates all fields of NodeInfo with Set from the
values of NodeConf.
*/
func (node *NodeInfo) SetFrom(n *NodeConf) {
	setWrap := func(entr *Entry, val string, nameArg string) {
		entr.Set(val)
	}
	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
		entr.SetSlice(val)
	}
	recursiveSetter(n, node, "", setWrap, setSliceWrap)
}

/*
Populates all fields of NodeInfo with SetAlt from the
values of NodeConf. The string profileName is used to
determine from which source/NodeInfo the entry came
from.
*/
func (node *NodeInfo) SetAltFrom(n *NodeConf, profileName string) {
	recursiveSetter(n, node, profileName, (*Entry).SetAlt, (*Entry).SetAltSlice)
}

/*
Populates all fields of NodeInfo with SetDefault from the
values of NodeConf.
*/
func (node *NodeInfo) SetDefFrom(n *NodeConf) {
	setWrap := func(entr *Entry, val string, nameArg string) {
		entr.SetDefault(val)
	}
	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
		entr.SetDefaultSlice(val)
	}
	recursiveSetter(n, node, "", setWrap, setSliceWrap)
}

func SetDefFrom(source, target interface{}) {
	setWrap := func(entr *Entry, val string, nameArg string) {
		entr.SetDefault(val)
	}
	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
		entr.SetDefaultSlice(val)
	}
	recursiveSetter(source, target, "", setWrap, setSliceWrap)

}

/*
Abstract function which populates a NodeInfo from a NodeConf via
setter functions. Panics if other type than string, []string *ptr is used in NodeConf.
*/
func recursiveSetter(source, target interface{}, nameArg string, setter func(*Entry, string, string),
	setterSlice func(*Entry, []string, string)) {
	sourceValue := reflect.ValueOf(source)
	targetType := reflect.TypeOf(target)
	targetValue := reflect.ValueOf(target)
	if targetValue.Elem().Kind() == reflect.Struct && sourceValue.Elem().Kind() == reflect.Struct {
		for i := 0; i < targetType.Elem().NumField(); i++ {
			sourceValueMatched := sourceValue.Elem().FieldByName(targetType.Elem().Field(i).Name)
			if sourceValueMatched.IsValid() {
				if targetValue.Elem().Field(i).Type() == reflect.TypeOf(Entry{}) {
					// get the fields which are part of the struct
					switch sourceValueMatched.Type() {
					case reflect.TypeOf(""):
						setter(targetValue.Elem().Field(i).Addr().Interface().(*Entry), sourceValueMatched.String(), nameArg)
					case reflect.TypeOf([]string{}):
						setterSlice(targetValue.Elem().Field(i).Addr().Interface().(*Entry), sourceValueMatched.Interface().([]string), nameArg)
					default:
						panic(fmt.Errorf("can't convert an Entry to %s", targetValue.Elem().Field(i).Type()))
					}
				} else if sourceValueMatched.Kind() == reflect.Ptr {
					// if we get a pointer, initialize if empty and then have a recursive call
					if targetValue.Elem().Field(i).IsZero() {
						targetValue.Elem().Field(i).Set(reflect.New(targetType.Elem().Field(i).Type.Elem()))
					}
					recursiveSetter(sourceValueMatched.Interface(), targetValue.Elem().Field(i).Interface(), nameArg, setter, setterSlice)
				} else if sourceValueMatched.Type().Kind() == reflect.Map {
					if targetValue.Elem().Field(i).IsZero() {
						targetValue.Elem().Field(i).Set(reflect.MakeMap(targetType.Elem().Field(i).Type))
					}
					// delete a ap element which is only in the target
					if targetValue.Elem().Field(i).Len() > 0 && targetValue.Elem().Field(i).Len() < 0 {
						sourceIter := sourceValueMatched.MapRange()
						targetIter := targetValue.Elem().Field(i).MapRange()
						for targetIter.Next() {
							sameKey := false
							for sourceIter.Next() {
								if sourceIter.Key() == targetIter.Key() {
									sameKey = true
								}
							}
							if !sameKey {
								targetValue.Elem().Field(i).SetMapIndex(targetIter.Key(), reflect.Value{})
							}
						}
					}
					sourceIter := sourceValueMatched.MapRange()
					if sourceValueMatched.Type().Elem() == reflect.TypeOf("") {
						// go over a simple map with strings
						for sourceIter.Next() {
							if !targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).IsValid() {
								newEntr := new(Entry)
								setter(newEntr, sourceIter.Value().String(), nameArg)
								targetValue.Elem().Field(i).SetMapIndex(sourceIter.Key(), reflect.ValueOf(newEntr))
							}
						}
					} else {
						// now the complicated map which contains pointers to objects
						for sourceIter.Next() {
							if !targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).IsValid() {
								newPtr := reflect.New(targetType.Elem().Field(i).Type.Elem().Elem())
								targetValue.Elem().Field(i).SetMapIndex(sourceIter.Key(), newPtr)
							}
							recursiveSetter(sourceIter.Value().Interface(), targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).Interface(), nameArg, setter, setterSlice)

						}
					}
				}
			}
		}
	}

}

/*
Flattens out a NodeConf, which means if there are no explicit values in *IpmiConf
or *KernelConf, these pointer will set to nil. This will remove something like
ipmi: {} from nodes.conf
*/
func (info *NodeConf) Flatten() {
	recursiveFlatten(info)
}
func recursiveFlatten(strct interface{}) {
	confType := reflect.TypeOf(strct)
	confVal := reflect.ValueOf(strct)
	for j := 0; j < confType.Elem().NumField(); j++ {
		if confVal.Elem().Field(j).Type().Kind() == reflect.Ptr && !confVal.Elem().Field(j).IsNil() {
			// iterate now over the ptr fields
			setToNil := true
			nestedType := reflect.TypeOf(confVal.Elem().Field(j).Interface())
			nestedVal := reflect.ValueOf(confVal.Elem().Field(j).Interface())
			for i := 0; i < nestedType.Elem().NumField(); i++ {
				if nestedType.Elem().Field(i).Type.Kind() == reflect.String &&
					nestedVal.Elem().Field(i).Interface().(string) != "" {
					setToNil = false
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf([]string{}) &&
					len(nestedVal.Elem().Field(i).Interface().([]string)) != 0 {
					setToNil = false
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf(map[string]string{}) &&
					len(nestedVal.Elem().Field(i).Interface().(map[string]string)) != 0 {
					setToNil = false
				}
			}
			if setToNil {
				confVal.Elem().Field(j).Set(reflect.Zero(confVal.Elem().Field(j).Type()))
			}
		}
	}
}

/*
Create a string slice, where every element represents a yaml entry, used for node/profile edit
in order to get a summary of all available elements
*/
func UnmarshalConf(obj interface{}, excludeList []string) (lines []string) {
	objType := reflect.TypeOf(obj)
	// now iterate of every field
	for i := 0; i < objType.NumField(); i++ {
		if objType.Field(i).Tag.Get("comment") != "" {
			if ymlStr, ok := getYamlString(objType.Field(i), excludeList); ok {
				lines = append(lines, ymlStr...)
			}
		}
		if objType.Field(i).Type.Kind() == reflect.Ptr && objType.Field(i).Tag.Get("yaml") != "" {
			typeLine := objType.Field(i).Tag.Get("yaml")
			if len(strings.Split(typeLine, ",")) > 1 {
				typeLine = strings.Split(typeLine, ",")[0] + ":"
			}
			lines = append(lines, typeLine)
			nestedLine := UnmarshalConf(reflect.New(objType.Field(i).Type.Elem()).Elem().Interface(), excludeList)
			for _, ln := range nestedLine {
				lines = append(lines, "  "+ln)
			}
		} else if objType.Field(i).Type.Kind() == reflect.Map && objType.Field(i).Type.Elem().Kind() == reflect.Ptr {
			typeLine := objType.Field(i).Tag.Get("yaml")
			if len(strings.Split(typeLine, ",")) > 1 {
				typeLine = strings.Split(typeLine, ",")[0] + ":"
			}
			lines = append(lines, typeLine, "  element:")
			nestedLine := UnmarshalConf(reflect.New(objType.Field(i).Type.Elem().Elem()).Elem().Interface(), excludeList)
			for _, ln := range nestedLine {
				lines = append(lines, "    "+ln)
			}
		}
	}
	return lines
}

/*
Get the string of the yaml tag
*/

func getYamlString(myType reflect.StructField, excludeList []string) ([]string, bool) {
	ymlStr := myType.Tag.Get("yaml")
	if len(strings.Split(ymlStr, ",")) > 1 {
		ymlStr = strings.Split(ymlStr, ",")[0]
	}
	if util.InSlice(excludeList, ymlStr) {
		return []string{""}, false
	} else if myType.Tag.Get("comment") == "" && myType.Type.Kind() == reflect.String {
		return []string{""}, false
	}
	if myType.Type.Kind() == reflect.String {
		fieldType := myType.Tag.Get("type")
		if fieldType == "" {
			fieldType = "string"
		}
		ymlStr += ": " + fieldType
		return []string{ymlStr}, true
	} else if myType.Type == reflect.TypeOf([]string{}) {
		return []string{ymlStr + ":", "  - string"}, true
	} else if myType.Type == reflect.TypeOf(map[string]string{}) {
		return []string{ymlStr + ":", "  key: value"}, true
	} else if myType.Type.Kind() == reflect.Ptr {
		return []string{ymlStr + ":"}, true
	}
	return []string{ymlStr}, true
}

/*
Set the field of the NodeConf with the given lopt name, returns true if the
field was found. String slices must be comma separated. Network must have the form
net.$NETNAME.lopt or netname.$NETNAME.lopt
*/
func (nodeConf *NodeConf) SetLopt(lopt string, value string) (found bool) {
	found = false
	nodeInfoType := reflect.TypeOf(nodeConf)
	nodeInfoVal := reflect.ValueOf(nodeConf)
	// try to find the normal fields, networks come later
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		if nodeInfoType.Elem().Field(i).Tag.Get("lopt") == lopt {
			if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.String {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s\n",
					lopt, nodeInfoType.Elem().Field(i).Name, value)
				confVal := nodeInfoVal.Elem().Field(i).Addr().Interface().(*string)
				*confVal = value
				found = true
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf([]string{}) {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s\n",
					lopt, nodeInfoType.Elem().Field(i).Name, value)
				confVal := nodeInfoVal.Elem().Field(i).Addr().Interface().(*[]string)
				*confVal = strings.Split(value, ",")
				found = true
			}
		}
	}
	// check network
	loptSlice := strings.Split(lopt, ".")
	wwlog.Debug("Trying to get network out of %s\n", loptSlice)
	if !found && len(loptSlice) == 3 && (loptSlice[0] == "net" || loptSlice[0] == "network" || loptSlice[0] == "netname") {
		if nodeConf.NetDevs == nil {
			nodeConf.NetDevs = make(map[string]*NetDevs)
		}
		if nodeConf.NetDevs[loptSlice[1]] == nil {
			nodeConf.NetDevs[loptSlice[1]] = new(NetDevs)
		}
		netInfoType := reflect.TypeOf(nodeConf.NetDevs[loptSlice[1]])
		netInfoVal := reflect.ValueOf(nodeConf.NetDevs[loptSlice[1]])
		for i := 0; i < netInfoVal.Elem().NumField(); i++ {
			if netInfoType.Elem().Field(i).Tag.Get("lopt") == loptSlice[2] {
				if netInfoType.Elem().Field(i).Type.Kind() == reflect.String {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s\n",
						lopt, loptSlice[1], netInfoType.Elem().Field(i).Name, value)
					confVal := netInfoVal.Elem().Field(i).Addr().Interface().(*string)
					*confVal = value
					found = true
				} else if netInfoType.Elem().Field(i).Type == reflect.TypeOf([]string{}) {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s\n",
						lopt, loptSlice[1], netInfoType.Elem().Field(i).Name, value)
					confVal := netInfoVal.Elem().Field(i).Addr().Interface().(*[]string)
					*confVal = strings.Split(value, ",")
					found = true
				}
			}
		}
	}
	return found
}
