package node

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

/**********
 *
 * Filters
 *
 *********/
/*
Filter a given slice of NodeInfo against a given
regular expression
*/
func FilterByName(set []NodeInfo, searchList []string) []NodeInfo {
	var ret []NodeInfo
	unique := make(map[string]NodeInfo)

	if len(searchList) > 0 {
		for _, search := range searchList {
			for _, entry := range set {
				b, _ := regexp.MatchString("^"+search+"$", entry.Id.Get())
				if b {
					unique[entry.Id.Get()] = entry
				}
			}
		}
		for _, n := range unique {
			ret = append(ret, n)
		}
	} else {
		ret = set
	}

	return ret
}

/**********
 *
 * Sets
 *
 *********/

/*
Set value. If argument is 'UNDEF', 'DELETE',
'UNSET" or '--' the value is removed.
N.B. the '--' might never ever happen as '--'
is parsed out by cobra
*/
func (ent *Entry) Set(val string) {
	if val == "" {
		return
	}

	if val == "UNDEF" || val == "DELETE" || val == "UNSET" || val == "--" || val == "nil" {
		wwlog.Debug("Removing value for %v\n", *ent)
		ent.value = []string{""}
	} else {
		ent.value = []string{val}
	}
}

/*
Set bool
*/
func (ent *Entry) SetB(val bool) {
	if val {
		ent.value = []string{"true"}
	} else {
		ent.value = []string{"false"}
	}
}

func (ent *Entry) SetSlice(val []string) {
	if len(val) == 0 {
		return
	}
	if val[0] == "UNDEF" || val[0] == "DELETE" || val[0] == "UNSET" || val[0] == "--" {
		ent.value = []string{}
	} else {
		ent.value = val
	}
}

/*
Set alternative value
*/
func (ent *Entry) SetAlt(val string, from string) {
	if val == "" {
		return
	}
	ent.altvalue = []string{val}
	ent.from = from
}

/*
Sets alternative bool
*/
func (ent *Entry) SetAltB(val bool, from string) {
	if val {
		ent.altvalue = []string{"true"}
		ent.from = from
	} else {
		ent.altvalue = []string{"false"}
		ent.from = from
	}
}

/*
Sets alternative slice
*/
func (ent *Entry) SetAltSlice(val []string, from string) {
	if len(val) == 0 {
		return
	}
	ent.altvalue = val
	ent.from = from
}

/*
Sets the default value of an entry.
*/
func (ent *Entry) SetDefault(val string) {
	if val == "" {
		return
	}
	ent.def = []string{val}

}

/*
Set the default entry as slice
*/
func (ent *Entry) SetDefaultSlice(val []string) {
	if len(val) == 0 {
		return
	}
	ent.def = val

}

/*
Remove a elemnt from a slice
*/
func (ent *Entry) SliceRemoveElement(val string) {
	util.SliceRemoveElement(ent.value, val)
}

/**********
*
* Gets
*
*********/
/*
Gets the the entry of the value in folowing order
* node value if set
* profile value if set
* default value if set
*/
func (ent *Entry) Get() string {
	if len(ent.value) != 0 {
		return ent.value[0]
	}
	if len(ent.altvalue) != 0 {
		return ent.altvalue[0]
	}
	if len(ent.def) != 0 {
		return ent.def[0]
	}
	return ""
}

/*
Get the bool value of an entry.
*/
func (ent *Entry) GetB() bool {
	if len(ent.value) == 0 || ent.value[0] == "false" || ent.value[0] == "no" {
		if len(ent.altvalue) == 0 || ent.altvalue[0] == "false" || ent.altvalue[0] == "no" {
			return false
		}
		return false
	}
	return true
}

/*
Returns a string slice created from a comma seperated list of the value.
*/
func (ent *Entry) GetSlice() []string {
	var retval []string
	if len(ent.value) != 0 {
		return ent.value
	}
	if len(ent.altvalue) != 0 {
		return ent.altvalue
	}
	if len(ent.def) != 0 {
		return ent.def
	}
	return retval
}

/*
Get the real value, not the alternative of default one.
*/
func (ent *Entry) GetReal() string {
	if len(ent.value) == 0 {
		return ""
	}
	return ent.value[0]
}

/*
Get the real value, not the alternative of default one.
*/
func (ent *Entry) GetRealSlice() []string {
	if len(ent.value) == 0 {
		return []string{}
	}
	return ent.value
}

/*
true if the entry has set a real value, else false.
*/
func (ent *Entry) GotReal() bool {
	return len(ent.value) != 0
}

/**********
 *
 * Misc
 *
 *********/

/*
Returns the value of Entry if it was defined set or
alternative is presend. Default value is in '()'. If
nothing is defined '--' is returned.
*/
func (ent *Entry) Print() string {
	if len(ent.value) != 0 {
		return strings.Join(ent.value, ",")
	}
	if len(ent.altvalue) != 0 {
		return strings.Join(ent.altvalue, ",")
	}
	if len(ent.def) != 0 {
		return "(" + strings.Join(ent.def, ",") + ")"
	}
	return "--"
}

/*
Was used for combined stringSlice

func (ent *Entry) PrintComb() string {
	if ent.value != "" && ent.altvalue != "" {
		return "[" + ent.value + "," + ent.altvalue + "]"
	}
	return ent.Print()
}
*/

/*
same as GetB()
*/
func (ent *Entry) PrintB() string {
	if len(ent.value) != 0 || len(ent.altvalue) != 0 {
		return fmt.Sprintf("%t", ent.GetB())
	}
	return fmt.Sprintf("(%t)", ent.GetB())
}

/*
Returns SUPERSEDED if value was set per node or
per profile. Else -- is returned.
*/
func (ent *Entry) Source() string {
	if len(ent.value) != 0 && len(ent.altvalue) != 0 {
		return "SUPERSEDED"
		//return fmt.Sprintf("[%s]", ent.from)
	} else if ent.from == "" {
		return "--"
	}
	return ent.from
}

/*
Check if value was defined.
*/
func (ent *Entry) Defined() bool {
	if len(ent.value) != 0 {
		return true
	}
	if len(ent.altvalue) != 0 {
		return true
	}
	if len(ent.def) != 0 {
		return true
	}
	return false
}

/*
Set the Entry trough an interface by trying to cast the interface
*/
func SetEntry(entryPtr interface{}, val interface{}) {
	valKind := reflect.TypeOf(val)
	if reflect.TypeOf(entryPtr) == reflect.TypeOf((*Entry)(nil)) {
		entry := entryPtr.(*Entry)
		if valKind.Kind() == reflect.String {
			entry.Set(val.(string))
		} else if valKind.Kind() == reflect.Slice {
			if valKind.Elem().Kind() == reflect.String {
				entry.SetSlice(val.([]string))
			} else {
				panic("Got unknown slice type")
			}
		}
	} else if reflect.TypeOf(entryPtr) == reflect.TypeOf((*[]string)(nil)) {
		entry := entryPtr.(*[]string)
		if valKind.Kind() == reflect.String {
			// most likely we got a comma seperated string slice
			*entry = strings.Split(val.(string), ",")
		} else if valKind.Kind() == reflect.Slice {
			if valKind.Elem().Kind() == reflect.String {
				*entry = val.([]string)
			} else {
				panic("Got unknown slice type")
			}
		}
	} else {
		panic(fmt.Sprintf("Can't convert %s to *node.Entry to call Set\n", reflect.TypeOf(entryPtr)))
	}

}

/*
Add an entry in a map
*/
func addEntry(entryMapInt interface{}, val interface{}) {
	if reflect.TypeOf(entryMapInt) == reflect.TypeOf((*map[string]*Entry)(nil)) {
		if reflect.ValueOf(entryMapInt).Elem().IsNil() {
			newMap := make(map[string]*Entry)
			entryMapInt = &newMap
		}
		entryMap := entryMapInt.(*map[string]*Entry)
		str, ok := (val).(string)
		if !ok {
			panic("AddEntry must be called with string value")
		}
		for _, token := range strings.Split(str, ",") {
			keyVal := strings.Split(token, "=")
			if len(keyVal) == 2 {
				_, mapOk := (*entryMap)[keyVal[0]]
				if !mapOk {
					(*entryMap)[keyVal[0]] = new(Entry)
				}
				(*entryMap)[keyVal[0]].Set(keyVal[1])
			}
		}
	} else {
		panic(fmt.Sprintf("Do not know how to add %v to %v\n", val, entryMapInt))
	}
}

/*
Del an entry in a map
*/
func delEntry(entryMapInt interface{}, val interface{}) {
	if reflect.TypeOf(entryMapInt) == reflect.TypeOf((*map[string]*Entry)(nil)) {
		entryMap := entryMapInt.(*map[string]*Entry)
		str, ok := (val).(string)
		if !ok {
			panic("DelEntry must be called with string value")
		}
		for _, token := range strings.Split(str, ",") {
			delete(*entryMap, token)
		}
	} else {
		panic(fmt.Sprintf("Do not know how to del %v to %v\n", val, entryMapInt))
	}

}

/*
Call SetEntry for given field (NodeInfo).
*/
func (node *NodeInfo) SetField(fieldName string, val interface{}) {
	field := reflect.ValueOf(node).Elem().FieldByName(fieldName)
	if field.IsValid() {
		if field.Addr().Type() == reflect.TypeOf((*Entry)(nil)) {
			SetEntry(field.Addr().Interface(), val)
		} else if field.Addr().Type() == reflect.TypeOf((*[]string)(nil)) {
			SetEntry(field.Addr().Interface(), val)
		}
		/*
			else if field.Addr().Kind() == reflect.Map {
				fmt.Println(field.Addr())
			} else {
				//fmt.Println("Not working field.Addr().Kind():", field.Addr().Type())
				// is most likely NetDevEntry, ignore it
			}
		*/
	} else {
		fieldNames := strings.Split(fieldName, ".")
		if len(fieldNames) >= 2 {
			if fieldNames[0] == "del" || fieldNames[0] == "add" {
				fieldMap := reflect.ValueOf(node).Elem().FieldByName(fieldNames[1])
				if fieldMap.IsValid() {
					if fieldNames[0] == "del" {
						delEntry(fieldMap.Addr().Interface(), val)
					} else if fieldNames[0] == "add" {
						addEntry(fieldMap.Addr().Interface(), val)
					}
				} else {
					panic(fmt.Sprintf("invalid del/add operation with name %s called, field %s does not exists\n", fieldName, fieldNames[0]))
				}
			} else {
				nestedField := reflect.ValueOf(node).Elem().FieldByName(fieldNames[0])
				if nestedField.IsValid() {
					switch nestedField.Addr().Type() {
					case reflect.TypeOf((**KernelEntry)(nil)):
						entry := nestedField.Addr().Interface().(**KernelEntry)
						(*entry).SetField(strings.Join(fieldNames[1:], "."), val)
					case reflect.TypeOf((**IpmiEntry)(nil)):
						entry := nestedField.Addr().Interface().(**IpmiEntry)
						(*entry).SetField(strings.Join(fieldNames[1:], "."), val)
					case reflect.TypeOf((*map[string]*NetDevEntry)(nil)):
						if len(fieldNames) >= 3 {
							entryMap := nestedField.Addr().Interface().(*map[string]*NetDevEntry)
							if myVal, ok := (*entryMap)[fieldNames[1]]; ok {
								myVal.SetField(strings.Join(fieldNames[2:], "."), val)
							} else {
								var newEntry NetDevEntry
								(*entryMap)[fieldNames[1]] = &newEntry
								newEntry.SetField(strings.Join(fieldNames[2:], "."), val)
							}
						}
					default:
						panic(fmt.Sprintf("not implemented type %v\n", nestedField.Addr().Type()))
					}
				} else {
					panic(fmt.Sprintf("field %s is not a nested type of %s", fieldNames[0], fieldName))
				}
			}
		} else {
			panic(fmt.Sprintf("field %s does not exists in node.NodeInfo\n", fieldName))
		}
	}

}

/*
Call SetEntry for given field (KernelEntry)
*/
func (node *KernelEntry) SetField(fieldName string, val interface{}) {
	field := reflect.ValueOf(node).Elem().FieldByName(fieldName)
	if field.IsValid() {
		SetEntry(field.Addr().Interface(), val)
	} else {
		valFields := strings.Split(fieldName, ".")
		field = reflect.ValueOf(node).Elem().FieldByName(valFields[1])
		if field.IsValid() && len(valFields) == 2 && valFields[0] == "add" {
			addEntry(field.Addr().Interface(), val)
		} else if field.IsValid() && len(valFields) == 2 && valFields[0] == "del" {
			delEntry(field.Addr().Interface(), val)
		} else {
			panic(fmt.Sprintf("field %s does not exists in node.NetDevEntry\n", fieldName))
		}
	}
}

/*
Call SetEntry for given field (ImpiEntry)
*/
func (node *IpmiEntry) SetField(fieldName string, val interface{}) {
	field := reflect.ValueOf(node).Elem().FieldByName(fieldName)
	if field.IsValid() {
		SetEntry(field.Addr().Interface(), val)
	} else {
		valFields := strings.Split(fieldName, ".")
		field = reflect.ValueOf(node).Elem().FieldByName(valFields[1])
		if field.IsValid() && len(valFields) == 2 && valFields[0] == "add" {
			addEntry(field.Addr().Interface(), val)
		} else if field.IsValid() && len(valFields) == 2 && valFields[0] == "del" {
			delEntry(field.Addr().Interface(), val)
		} else {
			panic(fmt.Sprintf("field %s does not exists in node.NetDevEntry\n", fieldName))
		}
	}
}

/*
Call SetEntry for given field (NetDevEntry)
*/
func (node *NetDevEntry) SetField(fieldName string, val interface{}) {
	field := reflect.ValueOf(node).Elem().FieldByName(fieldName)
	if field.IsValid() {
		SetEntry(field.Addr().Interface(), val)
	} else {
		valFields := strings.Split(fieldName, ".")
		field = reflect.ValueOf(node).Elem().FieldByName(valFields[1])
		if field.IsValid() && len(valFields) == 2 && valFields[0] == "add" {
			addEntry(field.Addr().Interface(), val)
		} else if field.IsValid() && len(valFields) == 2 && valFields[0] == "del" {
			delEntry(field.Addr().Interface(), val)
		} else {
			panic(fmt.Sprintf("field %s does not exists in node.NetDevEntry\n", fieldName))
		}
	}
}

/*
Get all names of the fields in the given struct (recursive)
and create a map[name of struct field]*string if the the field
of the struct bears the comment tag.
*/
func GetOptionsMap(theStruct interface{}) map[string]*string {
	optionsMap := make(map[string]*string)
	structVal := reflect.ValueOf(theStruct)
	structTyp := structVal.Type()
	for i := 0; i < structVal.NumField(); i++ {
		field := structTyp.Field(i)
		if field.Type.Kind() == reflect.Struct {
			subStruct := GetOptionsMap(field)
			for key, val := range subStruct {
				optionsMap[key] = val
			}
		} else if field.Tag.Get("comment") != "" {
			optionsMap[field.Name] = new(string)
		}

	}
	return optionsMap
}

type CobraCommand struct {
	*cobra.Command
}

/*
Get all names of the fields in the given struct (recursive)
and create a map[name of struct field]*string if the the field
of the struct bears the comment tag.
*/
func (baseCmd *CobraCommand) CreateFlags(theStruct interface{}, excludeList []string) map[string]*string {
	optionsMap := make(map[string]*string)
	structVal := reflect.ValueOf(theStruct)
	structTyp := structVal.Type()
	for i := 0; i < structVal.NumField(); i++ {
		field := structTyp.Field(i)
		//fmt.Printf("%s: field.Kind() == %s\n", field.Name, field.Type.Kind())
		if field.Type.Kind() == reflect.Ptr {
			a := structVal.Field(i).Elem().Interface()
			subStruct := baseCmd.CreateFlags(a, excludeList)
			for key, val := range subStruct {
				optionsMap[field.Name+"."+key] = val
			}

		} else if field.Type.Kind() == reflect.Map {
			// check the type of map
			mapType := field.Type.Elem()
			if mapType.Kind() == reflect.Ptr {
				//a := reflect.ValueOf((mapType.Elem())) node.NetDevs
				subMap := baseCmd.CreateFlags(reflect.New(mapType.Elem()).Elem().Interface(), excludeList)
				for key, val := range subMap {
					optionsMap[field.Name+"."+key] = val
				}
				if mapType == reflect.TypeOf((*NetDevs)(nil)) {
					// set the option for the network name here
					var netName string
					optionsMap[field.Name] = &netName
					baseCmd.PersistentFlags().StringVarP(&netName,
						"netname", "n", "", "Define the network name to configure")
				}
			} else if mapType.Kind() == reflect.String {
				if field.Tag.Get("lopt") != "" {
					var addPair string
					optionsMap["add"+"."+field.Name] = &addPair
					baseCmd.PersistentFlags().StringVarP(&addPair,
						field.Tag.Get("lopt")+"add", "", "", "Add key/value pair to "+field.Tag.Get("comment"))
					var delPair string
					optionsMap["del"+"."+field.Name] = &delPair
					baseCmd.PersistentFlags().StringVarP(&delPair,
						field.Tag.Get("lopt")+"del", "", "", "Delete key/value pair to "+field.Tag.Get("comment"))
				}
			} else {
				// TODO: implement handling of string maps
				wwlog.Warn("handling of %v not implemented\n", field.Type)
			}

		} else if field.Tag.Get("comment") != "" && !util.InSlice(excludeList, field.Tag.Get("lopt")) {
			var newStr string
			optionsMap[field.Name] = &newStr
			if field.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringVarP(&newStr,
					field.Tag.Get("lopt"),
					field.Tag.Get("sopt"),
					field.Tag.Get("default"),
					field.Tag.Get("comment"))
			} else if !util.InSlice(excludeList, field.Tag.Get("lopt")) {
				baseCmd.PersistentFlags().StringVar(&newStr,
					field.Tag.Get("lopt"),
					field.Tag.Get("default"),
					field.Tag.Get("comment"))

			}
		}

	}
	return optionsMap
}

/*
Helper function which gets the lopt of a given interface
*/
func GetLoptOf(myStruct interface{}, name string) string {
	retStr := ""
	if reflect.TypeOf(myStruct).Kind() != reflect.Struct {
		return retStr
	}
	myType := reflect.TypeOf(myStruct)
	field, ok := myType.FieldByName(name)
	if ok {
		retStr = field.Tag.Get("lopt")
	}
	return retStr

}

/*
Returns a translation map of field name and its associated lopt.
*/
func GetloptMap(myStruct interface{}) map[string]string {
	retMap := make(map[string]string)
	if reflect.TypeOf(myStruct).Kind() != reflect.Struct {
		return retMap
	}
	structType := reflect.TypeOf(myStruct)
	for i := 0; i < structType.NumField(); i++ {
		retMap[structType.Field(i).Name] = structType.Field(i).Name
		lopt := structType.Field(i).Tag.Get("lopt")
		if lopt != "" {
			retMap[structType.Field(i).Name] = lopt
		}
	}
	return retMap
}
