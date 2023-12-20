package node

import (
	"bytes"
	"encoding/gob"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwtype"
)

type sortByName []NodeConf

func (a sortByName) Len() int           { return len(a) }
func (a sortByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortByName) Less(i, j int) bool { return a[i].id < a[j].id }

/**********
 *
 * Filters
 *
 *********/

/*
Filter a given slice of NodeConf against a given
regular expression
*/
func FilterByName(set []NodeConf, searchList []string) []NodeConf {
	var ret []NodeConf
	unique := make(map[string]NodeConf)

	if len(searchList) > 0 {
		for _, search := range searchList {
			for _, entry := range set {
				if match, _ := regexp.MatchString("^"+search+"$", entry.id); match {
					unique[entry.id] = entry
				}
			}
		}
		for _, n := range unique {
			ret = append(ret, n)
		}
	} else {
		ret = set
	}

	sort.Sort(sortByName(ret))
	return ret
}

/*
Filter a given map of NodeConf against given regular expression.
*/
func FilterNodesByName(inputMap map[string]*NodeConf, searchList []string) (retMap map[string]*NodeConf) {
	retMap = map[string]*NodeConf{}
	if len(searchList) > 0 {
		for _, search := range searchList {
			for name, nConf := range inputMap {
				if match, _ := regexp.MatchString("^"+search+"$", name); match {
					retMap[name] = nConf
				}
			}
		}
	}
	return retMap
}

/*
Filter a given map of NodeConf against given regular expression.
*/
func FilterProfilesByName(inputMap map[string]*ProfileConf, searchList []string) (retMap map[string]*ProfileConf) {
	retMap = map[string]*ProfileConf{}
	if len(searchList) > 0 {
		for _, search := range searchList {
			for name, nConf := range inputMap {
				if match, _ := regexp.MatchString("^"+search+"$", name); match {
					retMap[name] = nConf
				}
			}
		}
	}
	return retMap
}

/*
Creates an NodeConf with the given id. Doesn't add it to the database
*/
func NewNode(id string) (nodeconf NodeConf) {
	nodeconf = EmptyNode()
	nodeconf.id = id
	return nodeconf
}

func EmptyNode() (nodeconf NodeConf) {
	nodeconf.Ipmi = new(IpmiConf)
	nodeconf.Ipmi.Tags = map[string]string{}
	nodeconf.Kernel = new(KernelConf)
	nodeconf.NetDevs = make(map[string]*NetDevs)
	nodeconf.Tags = map[string]string{}
	return nodeconf
}

/*
Creates a ProfileConf but doesn't add it to the database.
*/
func EmptyProfile() (profileconf ProfileConf) {
	profileconf.Ipmi = new(IpmiConf)
	profileconf.Ipmi.Tags = map[string]string{}
	profileconf.Kernel = new(KernelConf)
	profileconf.NetDevs = make(map[string]*NetDevs)
	profileconf.Tags = map[string]string{}
	return profileconf
}

/*
Flattens out a NodeConf, which means if there are no explicit values in *IpmiConf
or *KernelConf, these pointer will set to nil. This will remove something like
ipmi: {} from nodes.conf
*/
func (info *NodeConf) Flatten() {
	recursiveFlatten(info)
}

/*
Flattens out a ProfileConf, which means if there are no explicit values in *IpmiConf
or *KernelConf, these pointer will set to nil. This will remove something like
ipmi: {} from nodes.conf
*/
func (info *ProfileConf) Flatten() {
	recursiveFlatten(info)
}

// abstract flatten
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
				// wwlog.Debug("checking %s", nestedType.Elem().Field(i).Type.String())
				if nestedType.Elem().Field(i).Type.Kind() == reflect.String &&
					nestedVal.Elem().Field(i).Interface().(string) != "" &&
					nestedVal.Elem().Field(i).Interface().(string) != undef {
					setToNil = false
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf([]string{}) &&
					len(nestedVal.Elem().Field(i).Interface().([]string)) != 0 {
					setToNil = false
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf(map[string]string{}) &&
					len(nestedVal.Elem().Field(i).Interface().(map[string]string)) != 0 {
					setToNil = false
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf(net.IP{}) {
					val := nestedVal.Elem().Field(i).Interface().(net.IP)
					if len(val) != 0 && !val.IsUnspecified() {
						setToNil = false
					}
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf(wwtype.WWbool{}) {
					val := nestedVal.Elem().Field(i).Interface().(wwtype.WWbool)
					if !val.IsZero() {
						setToNil = false
					}
				}
			}
			if setToNil {
				confVal.Elem().Field(j).Set(reflect.Zero(confVal.Elem().Field(j).Type()))
			}
		} else if confType.Elem().Field(j).Anonymous {
			recursiveFlatten(confVal.Elem().Field(j).Addr().Interface())
		} else if confType.Elem().Field(j).IsExported() && confType.Elem().Field(j).Type.Kind() == reflect.String {
			if confVal.Elem().Field(j).Interface().(string) == undef {
				confVal.Elem().Field(j).SetString("")
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
Getters for unexported fields
*/

/*
Returns the id of the node
*/
func (node *NodeConf) Id() string {
	return node.id
}

/*
Returns the id of the profile
*/
func (node *ProfileConf) Id() string {
	return node.id
}

/*
Returns if the node is a valid in the database
*/
func (node *NodeConf) Valid() bool {
	return node.valid
}

/*
Check if the netdev is the primary one
*/
func (dev *NetDevs) Primary() bool {
	return dev.primary
}

// returns all negated elements which are marked with ! as prefix
// from a list
func negList(list []string) (ret []string) {
	for _, tok := range list {
		if strings.HasPrefix(tok, "~") {
			ret = append(ret, tok[1:])
		}
	}
	return
}

// clean a list from negated tokens
func cleanList(list []string) (ret []string) {
	neg := negList(list)
	for _, listTok := range list {
		notNegate := true
		for _, negTok := range neg {
			if listTok == negTok || listTok == "~"+negTok {
				notNegate = false
			}
		}
		if notNegate {
			ret = append(ret, listTok)
		}
	}
	return ret
}

// Clone (deep copy) a node via gob
func (src *NodeConf) Clone() (dst NodeConf, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err = enc.Encode(src)
	if err != nil {
		return
	}
	err = dec.Decode(&dst)
	if err != nil {
		return
	}
	dst.id = src.id
	dst.valid = src.valid
	return
}

// Clone (deep copy) a node via gob
func (src ProfileConf) Clone() (dst ProfileConf, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err = enc.Encode(src)
	if err != nil {
		return
	}
	err = dec.Decode(&dst)
	if err != nil {
		return
	}
	dst.id = src.id
	return
}
